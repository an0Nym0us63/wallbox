package bridge

import (
	"fmt"
	"strconv"

	"github.com/jagheterfredrik/wallbox-mqtt-bridge/app/wallbox"
)

type Entity struct {
	Component string
	Getter    func() string
	Setter    func(string)
	Config    map[string]string
}

func strToInt(val string) int {
	i, _ := strconv.Atoi(val)
	return i
}

func strToFloat(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func getEntities(w *wallbox.Wallbox) map[string]Entity {
	return map[string]Entity{
		"added_energy_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisState.ScheduleEnergy/1000) },
			Config: map[string]string{
				"name":                        "Active Added energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": "1",
			},
		},
		"green_energy_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisState.GreenEnergy/1000) },
			Config: map[string]string{
				"name":                        "Green energy active",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": "1",
			},
		},
		"grid_energy_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint((w.Data.RedisState.ScheduleEnergy-w.Data.RedisState.GreenEnergy)/1000) },
			Config: map[string]string{
				"name":                        "Active Grid Energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"added_range_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisState.AddedRange) },
			Config: map[string]string{
				"name":                        "Added Range active",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": "1",
			},
		},
		"charging_time_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisState.ChargingTime) },
			Config: map[string]string{
				"name":                        "Charging Time active",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": "1",
			},
		},
		"charging_speed_active": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisState.ChargingSpeed) },
			Config: map[string]string{
				"name":                        "Charging Speed active",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total",
				"suggested_display_precision": "1",
			},
		},
		"added_energy": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.AddedEnergy/1000) },
			Config: map[string]string{
				"name":                        "Added Energy",
				"device_class":                "distance",
				"unit_of_measurement":         "km",
				"state_class":                 "total",
				"suggested_display_precision": "1",
				"icon":                        "mdi:map-marker-distance",
			},
		},
		"added_range": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.AddedRange) },
			Config: map[string]string{
				"name":                        "Added range",
				"device_class":                "distance",
				"unit_of_measurement":         "km",
				"state_class":                 "total",
				"suggested_display_precision": "1",
				"icon":                        "mdi:map-marker-distance",
			},
		},
		"cable_connected": {
			Component: "binary_sensor",
			Getter:    func() string { return strconv.Itoa(w.CableConnected()) },
			Config: map[string]string{
				"name":         "Cable connected",
				"payload_on":   "1",
				"payload_off":  "0",
				"icon":         "mdi:ev-plug-type1",
				"device_class": "plug",
			},
		},
		"charging_enable": {
			Component: "switch",
			Setter:    func(val string) { w.SetChargingEnable(strToInt(val)) },
			Getter:    func() string { return strconv.Itoa(w.Data.SQL.ChargingEnable) },
			Config: map[string]string{
				"name":        "Charging enable",
				"payload_on":  "1",
				"payload_off": "0",
				"icon":        "mdi:ev-station",
			},
		},
		"charging_power": {
			Component: "sensor",
			Getter: func() string {
				return fmt.Sprint(w.Data.RedisM2W.Line1Power + w.Data.RedisM2W.Line2Power + w.Data.RedisM2W.Line3Power)
			},
			Config: map[string]string{
				"name":                        "Charging power",
				"device_class":                "power",
				"unit_of_measurement":         "W",
				"state_class":                 "measurement",
				"suggested_display_precision": "1",
			},
		},
		"cumulative_added_energy": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.CumulativeAddedEnergy/1000) },
			Config: map[string]string{
				"name":                        "Cumulative added energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "total_increasing",
				"suggested_display_precision": "1",
			},
		},
		"total_cost": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.TotalCost) },
			Config: map[string]string{
				"name":                        "Cost of charge session",
				"device_class":                "energy",
				"unit_of_measurement":         "€",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"grid_cost": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.EnergyCost*(w.Data.SQL.AddedEnergy-w.Data.SQL.GreenEnergy)/1000) },
			Config: map[string]string{
				"name":                        "Cost of grid",
				"device_class":                "energy",
				"unit_of_measurement":         "€",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"halo_brightness": {
			Component: "number",
			Setter:    func(val string) { w.SetHaloBrightness(strToInt(val)) },
			Getter:    func() string { return strconv.Itoa(w.Data.SQL.HaloBrightness) },
			Config: map[string]string{
				"name":                "Halo Brightness",
				"command_topic":       "~/set",
				"min":                 "0",
				"max":                 "100",
				"icon":                "mdi:brightness-percent",
				"unit_of_measurement": "%",
				"entity_category":     "config",
			},
		},
		"car_battery": {
			Component: "sensor",
			Setter:    func(val string) { w.SetCarBattery(strToFloat(val)) },
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.CarBattery/1000) },
			Config: map[string]string{
				"name":                        "Car Battery",
				"command_topic":               "~/set",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"car_consumption": {
			Component: "sensor",
			Setter:    func(val string) { w.SetCarConsumption(strToFloat(val)) },
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.CarConsumption/10) },
			Config: map[string]string{
				"name":                        "Car Consumption",
				"command_topic":               "~/set",
				"device_class":                "energy",
				"unit_of_measurement":         "kWh/100",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"energy_cost": {
			Component: "sensor",
			Setter:    func(val string) { w.SetEnergyCost(strToFloat(val)) },
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.EnergyCost) },
			Config: map[string]string{
				"name":                        "Energy Cost",
				"command_topic":               "~/set",
				"device_class":                "energy",
				"unit_of_measurement":         "€/kWh",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"charging_time": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.ChargingTime) },
			Config: map[string]string{
				"name":                        "Effective Charging Time",
				"device_class":                "energy",
				"unit_of_measurement":         "s",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"green_energy": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.SQL.GreenEnergy/1000) },
			Config: map[string]string{
				"name":                        "Added Green Energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"grid_energy": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint((w.Data.SQL.AddedEnergy-w.Data.SQL.GreenEnergy)/1000) },
			Config: map[string]string{
				"name":                        "Added Grid Energy",
				"device_class":                "energy",
				"unit_of_measurement":         "Wh",
				"state_class":                 "measurement",
				"suggested_display_precision": "2",
			},
		},
		"lock": {
			Component: "lock",
			Setter:    func(val string) { w.SetLocked(strToInt(val)) },
			Getter:    func() string { return strconv.Itoa(w.Data.SQL.Lock) },
			Config: map[string]string{
				"name":           "Lock",
				"payload_lock":   "1",
				"payload_unlock": "0",
				"state_locked":   "1",
				"state_unlocked": "0",
				"command_topic":  "~/set",
			},
		},
		"max_charging_current": {
			Component: "number",
			Setter:    func(val string) { w.SetMaxChargingCurrent(strToInt(val)) },
			Getter:    func() string { return strconv.Itoa(w.Data.SQL.MaxChargingCurrent) },
			Config: map[string]string{
				"name":                "Max charging current",
				"command_topic":       "~/set",
				"min":                 "6",
				"max":                 strconv.Itoa(w.AvailableCurrent()),
				"unit_of_measurement": "A",
				"device_class":        "current",
			},
		},
		"status": {
			Component: "sensor",
			Getter:    w.EffectiveStatus,
			Config: map[string]string{
				"name": "Status",
			},
		},
		"start_time": {
			Component: "sensor",
			Getter:    func() string { return w.Data.SQL.StartTime },
			Config: map[string]string{
				"name": "Start Time",
			},
		},
		"end_time": {
			Component: "sensor",
			Getter:    func() string { return w.Data.SQL.EndTime },
			Config: map[string]string{
				"name": "End Time",
			},
		},
	}
}

func getDebugEntities(w *wallbox.Wallbox) map[string]Entity {
	return map[string]Entity{
		"control_pilot": {
			Component: "sensor",
			Getter:    w.ControlPilotStatus,
			Config: map[string]string{
				"name": "Control pilot",
			},
		},
		"m2w_status": {
			Component: "sensor",
			Getter:    func() string { return fmt.Sprint(w.Data.RedisM2W.ChargerStatus) },
			Config: map[string]string{
				"name": "M2W Status",
			},
		},
		"state_machine_state": {
			Component: "sensor",
			Getter:    w.StateMachineState,
			Config: map[string]string{
				"name": "State machine",
			},
		},
		"s2_open": {
			Component: "sensor",
			Getter:    func() string { return strconv.Itoa(w.Data.RedisState.S2open) },
			Config: map[string]string{
				"name": "S2 open",
			},
		},
	}
}
