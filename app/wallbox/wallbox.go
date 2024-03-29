package wallbox

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type DataCache struct {
	SQL struct {
		Lock                  int     `db:"lock"`
		ChargingEnable        int     `db:"charging_enable"`
		MaxChargingCurrent    int     `db:"max_charging_current"`
		HaloBrightness        int     `db:"halo_brightness"`
		CumulativeAddedEnergy float64 `db:"cumulative_added_energy"`
		AddedRange            float64 `db:"added_range"`
		ChargingTime          float64 `db:"charging_time"`
		GreenEnergy           float64 `db:"green_energy"`
		TotalCost             float64 `db:"total_cost"`
		CarConsumption        float64 `db:"car_consumption"`
		CarBattery            float64 `db:"car_battery"`
		EnergyCost            float64 `db:"energy_cost"`
		AddedEnergy           float64 `db:"energy_total"`
		StartTime             string  `db:"start_time"`
		EndTime               string  `db:"end_time"`
	}

	RedisState struct {
		SessionState   int     `redis:"session.state"`
		ControlPilot   int     `redis:"ctrlPilot"`
		S2open         int     `redis:"S2open"`
		ScheduleEnergy float64 `redis:"scheduleEnergy"`
		GreenEnergy    float64 `redis:"session.green_energy"`
		AddedRange     float64 `redis:"session.charged_range"`
		ChargingTime   float64 `redis:"session.charging_time"`
		ChargingSpeed  float64 `redis:"session.charging_speed"`
	}

	RedisM2W struct {
		ChargerStatus int     `redis:"tms.charger_status"`
		Line1Power    float64 `redis:"tms.line1.power_watt.value"`
		Line2Power    float64 `redis:"tms.line2.power_watt.value"`
		Line3Power    float64 `redis:"tms.line3.power_watt.value"`
	}
}

type Wallbox struct {
	redisClient *redis.Client
	sqlClient   *sqlx.DB
	Data        DataCache
	ChargerType string `db:"charger_type"`
}

func New() *Wallbox {
	var w Wallbox

	var err error
	w.sqlClient, err = sqlx.Connect("mysql", "root:fJmExsJgmKV7cq8H@tcp(127.0.0.1:3306)/wallbox")
	if err != nil {
		panic(err)
	}

	query := "select SUBSTRING_INDEX(part_number, '-', 1) AS charger_type from charger_info;"
	w.sqlClient.Get(&w, query)

	w.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &w
}

func getRedisFields(obj interface{}) []string {
	var result []string
	val := reflect.ValueOf(obj)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		result = append(result, field.Tag.Get("redis"))
	}

	return result
}

func (w *Wallbox) RefreshData() {
	ctx := context.Background()

	stateRes := w.redisClient.HMGet(ctx, "state", getRedisFields(w.Data.RedisState)...)
	if stateRes.Err() != nil {
		panic(stateRes.Err())
	}

	if err := stateRes.Scan(&w.Data.RedisState); err != nil {
		panic(err)
	}

	m2wRes := w.redisClient.HMGet(ctx, "m2w", getRedisFields(w.Data.RedisM2W)...)
	if m2wRes.Err() != nil {
		panic(m2wRes.Err())
	}

	if err := m2wRes.Scan(&w.Data.RedisM2W); err != nil {
		panic(err)
	}

	query := "SELECT " +
		"  `wallbox_config`.`charging_enable`," +
		"  `wallbox_config`.`lock`," +
		"  `wallbox_config`.`max_charging_current`," +
		"  `wallbox_config`.`halo_brightness`," +
		"  `power_outage_values`.`charged_energy` AS cumulative_added_energy," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    0," +
		"    `latest_session`.`total_cost`) AS total_cost," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    `active_session`.`charging_time`," +
		"    `latest_session`.`charging_time`) AS charging_time," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    0," +
		"    `latest_session`.`green_energy`) AS green_energy," +
		"  `first_energy`.`cost` AS energy_cost," +
		"  `first_car`.`consumption` AS car_consumption," +
		"  `first_car`.`battery` AS car_battery," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    'Session en cours'," +
		"    to_char(`latest_session`.`end_time`,'YYYY-MM-DD HH24:MI:SS')) AS end_time," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    to_char(`active_session`.`start_timestamp`,'YYYY-MM-DD HH24:MI:SS')," +
		"    to_char(`latest_session`.`start_time`,'YYYY-MM-DD HH24:MI:SS')) AS start_time," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    `active_session`.`charged_range`," +
		"    `latest_session`.`charged_range`) AS added_range," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    `active_session`.`energy_total`," +
		"    `latest_session`.`energy_total`) AS energy_total " +
		"FROM `wallbox_config`," +
		"    `active_session`," +
		"    `power_outage_values`," +
		"    (SELECT * FROM `session` ORDER BY `id` DESC LIMIT 1) AS latest_session," +
		"    (SELECT * FROM `energy` ORDER BY `id` ASC LIMIT 1) AS first_energy," +
		"    (SELECT * FROM `cars` ORDER BY `car_id` ASC LIMIT 1) AS first_car"
	w.sqlClient.Get(&w.Data.SQL, query)
}

func (w *Wallbox) SerialNumber() string {
	var serialNumber string
	w.sqlClient.Get(&serialNumber, "SELECT `serial_num` FROM charger_info")
	return serialNumber
}

func (w *Wallbox) UserId() string {
	var userId string
	w.sqlClient.QueryRow("SELECT `user_id` FROM `users` WHERE `user_id` != 1 ORDER BY `user_id` DESC LIMIT 1").Scan(&userId)
	return userId
}

func (w *Wallbox) SetCarConsumption(consumption float64) {
	w.sqlClient.MustExec("UPDATE `cars` SET `consumption`=? where car_id=1", consumption*10)
}

func (w *Wallbox) SetEnergyCost(cost float64) {
	w.sqlClient.MustExec("UPDATE `energy` SET `cost`=? where id=1", cost)
}

func (w *Wallbox) SetCarBattery(battery float64) {
	w.sqlClient.MustExec("UPDATE `cars` SET `battery`=? where car_id=1", battery*1000)
}

func (w *Wallbox) AvailableCurrent() int {
	var availableCurrent int
	w.sqlClient.QueryRow("SELECT `max_avbl_current` FROM `state_values` ORDER BY `id` DESC LIMIT 1").Scan(&availableCurrent)
	return availableCurrent
}

func sendToPosixQueue(path, data string) {
	pathBytes := append([]byte(path), 0)
	mq := mqOpen(pathBytes)

	event := []byte(data)
	eventPaddedBytes := append(event, bytes.Repeat([]byte{0x00}, 1024-len(event))...)

	mqTimedsend(mq, eventPaddedBytes)
	mqClose(mq)
}

func (w *Wallbox) SetLocked(lock int) {
	w.RefreshData()
	if lock == w.Data.SQL.Lock {
		return
	}
	if w.ChargerType == "CPB1" {
		w.sqlClient.MustExec("UPDATE `wallbox_config` SET `lock`=?", lock)
	} else if lock == 1 {
		sendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_LOGIN", "EVENT_REQUEST_LOCK")
	} else {
		userId := w.UserId()
		sendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_LOGIN", "EVENT_REQUEST_LOGIN#"+userId+".000000")
	}
}

func (w *Wallbox) SetChargingEnable(enable int) {
	w.RefreshData()
	if enable == w.Data.SQL.ChargingEnable {
		return
	}
	if enable == 1 {
		sendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_STATEMACHINE", "EVENT_REQUEST_USER_ACTION#1.000000")
	} else {
		sendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_STATEMACHINE", "EVENT_REQUEST_USER_ACTION#2.000000")
	}
}

func (w *Wallbox) SetMaxChargingCurrent(current int) {
	w.sqlClient.MustExec("UPDATE `wallbox_config` SET `max_charging_current`=?", current)
}

func (w *Wallbox) SetHaloBrightness(brightness int) {
	w.sqlClient.MustExec("UPDATE `wallbox_config` SET `halo_brightness`=?", brightness)
}

func (w *Wallbox) CableConnected() int {
	if w.Data.RedisM2W.ChargerStatus == 0 || w.Data.RedisM2W.ChargerStatus == 6 {
		return 0
	}
	return 1
}

func (w *Wallbox) EffectiveStatus() string {
	tmsStatus := w.Data.RedisM2W.ChargerStatus
	state := w.Data.RedisState.SessionState

	if override, ok := stateOverrides[state]; ok {
		tmsStatus = override
	}

	return wallboxStatusCodes[tmsStatus]
}

func (w *Wallbox) ControlPilotStatus() string {
	return fmt.Sprintf("%d: %s", w.Data.RedisState.ControlPilot, controlPilotStates[w.Data.RedisState.ControlPilot])
}

func (w *Wallbox) StateMachineState() string {
	return fmt.Sprintf("%d: %s", w.Data.RedisState.SessionState, stateMachineStates[w.Data.RedisState.SessionState])
}
