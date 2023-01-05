package main

import "strconv"

type Meters struct {
	Date          string
	HotWater      int
	ColdWater     int
	ElectricityT1 int
	ElectricityT2 int
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
