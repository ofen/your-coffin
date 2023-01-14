package main

import "strconv"

type Meters struct {
	Date          string `json:"date"`
	HotWater      int    `json:"how_water"`
	ColdWater     int    `json:"cold_water"`
	ElectricityT1 int    `json:"electricity_t1"`
	ElectricityT2 int    `json:"electricity_t2"`
}

func (t Meters) Sub(m *Meters) *Meters {
	return &Meters{
		Date:          t.Date,
		HotWater:      t.HotWater - m.HotWater,
		ColdWater:     t.ColdWater - m.ColdWater,
		ElectricityT1: t.ElectricityT1 - m.ElectricityT1,
		ElectricityT2: t.ElectricityT2 - m.ElectricityT2,
	}
}

func Rtom(row []interface{}) *Meters {
	m := &Meters{}

	m.Date = row[0].(string)
	m.HotWater, _ = strconv.Atoi(row[1].(string))
	m.ColdWater, _ = strconv.Atoi(row[2].(string))
	m.ElectricityT1, _ = strconv.Atoi(row[3].(string))
	m.ElectricityT2, _ = strconv.Atoi(row[4].(string))

	return m
}

func Mtor(m *Meters) []interface{} {
	return []interface{}{m.Date, m.HotWater, m.ColdWater, m.ElectricityT1, m.ElectricityT2}
}
