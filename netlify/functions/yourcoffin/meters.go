package main

import "strconv"

type meters struct {
	Date          string `json:"date"`
	HotWater      int    `json:"how_water"`
	ColdWater     int    `json:"cold_water"`
	ElectricityT1 int    `json:"electricity_t1"`
	ElectricityT2 int    `json:"electricity_t2"`
}

func (m meters) sub(m2 *meters) *meters {
	return &meters{
		Date:          m.Date,
		HotWater:      m.HotWater - m2.HotWater,
		ColdWater:     m.ColdWater - m2.ColdWater,
		ElectricityT1: m.ElectricityT1 - m2.ElectricityT1,
		ElectricityT2: m.ElectricityT2 - m2.ElectricityT2,
	}
}

func (m meters) toRow() []interface{} {
	return []interface{}{m.Date, m.HotWater, m.ColdWater, m.ElectricityT1, m.ElectricityT2}
}

func rowToMeters(row []interface{}) *meters {
	m := &meters{}

	m.Date = row[0].(string)
	m.HotWater, _ = strconv.Atoi(row[1].(string))
	m.ColdWater, _ = strconv.Atoi(row[2].(string))
	m.ElectricityT1, _ = strconv.Atoi(row[3].(string))
	m.ElectricityT2, _ = strconv.Atoi(row[4].(string))

	return m
}

func lastMeters() (*meters, error) {
	lastRows, err := gs.LastRow()
	if err != nil {
		return nil, err
	}

	return rowToMeters(lastRows), nil
}

func appendMeters(m *meters) error {
	return gs.AppendRow(m.toRow())
}

func listMeters() ([]*meters, error) {
	v, err := gs.Rows()
	if err != nil {
		return nil, err
	}

	res := make([]*meters, 0, len(v.Values))

	for _, vv := range v.Values {
		res = append(res, rowToMeters(vv))
	}

	return res, nil
}
