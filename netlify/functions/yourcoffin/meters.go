package main

import "strconv"

type Meters struct {
	Date          string `json:"date"`
	HotWater      int    `json:"how_water"`
	ColdWater     int    `json:"cold_water"`
	ElectricityT1 int    `json:"electricity_t1"`
	ElectricityT2 int    `json:"electricity_t2"`
}

func (m Meters) Sub(meters *Meters) *Meters {
	return &Meters{
		Date:          m.Date,
		HotWater:      m.HotWater - meters.HotWater,
		ColdWater:     m.ColdWater - meters.ColdWater,
		ElectricityT1: m.ElectricityT1 - meters.ElectricityT1,
		ElectricityT2: m.ElectricityT2 - meters.ElectricityT2,
	}
}

func (m Meters) ToRow() []interface{} {
	return []interface{}{m.Date, m.HotWater, m.ColdWater, m.ElectricityT1, m.ElectricityT2}
}

func RowToMeters(row []interface{}) *Meters {
	m := &Meters{}

	m.Date = row[0].(string)
	m.HotWater, _ = strconv.Atoi(row[1].(string))
	m.ColdWater, _ = strconv.Atoi(row[2].(string))
	m.ElectricityT1, _ = strconv.Atoi(row[3].(string))
	m.ElectricityT2, _ = strconv.Atoi(row[4].(string))

	return m
}
