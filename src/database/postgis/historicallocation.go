package postgis

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	gostErrors "github.com/geodan/gost/src/errors"
	"github.com/geodan/gost/src/sensorthings/entities"
	"github.com/geodan/gost/src/sensorthings/odata"
)

var hlMapping = map[string]string{"time": fmt.Sprintf("to_char(time at time zone 'UTC', '%s') as time", TimeFormat)}

// GetTotalHistoricalLocations returns the amount of HistoricalLocations in the database
func (gdb *GostDatabase) GetTotalHistoricalLocations() int {
	var count int
	sql := fmt.Sprintf("SELECT Count(*) from %s.historicallocation", gdb.Schema)
	gdb.Db.QueryRow(sql).Scan(&count)
	return count
}

// GetHistoricalLocation retireves a HistoricalLocation by id
func (gdb *GostDatabase) GetHistoricalLocation(id interface{}, qo *odata.QueryOptions) (*entities.HistoricalLocation, error) {
	intID, ok := ToIntID(id)
	if !ok {
		return nil, gostErrors.NewRequestNotFound(errors.New("HistoricalLocation does not exist"))
	}

	sql := fmt.Sprintf("select "+CreateSelectString(&entities.HistoricalLocation{}, qo, "", "", hlMapping)+" FROM %s.historicallocation where id = %v", gdb.Schema, intID)
	historicallocation, err := processHistoricalLocation(gdb.Db, sql, qo)
	if err != nil {
		return nil, err
	}

	return historicallocation, nil
}

// GetHistoricalLocations retrieves all historicallocations
func (gdb *GostDatabase) GetHistoricalLocations(qo *odata.QueryOptions) ([]*entities.HistoricalLocation, int, error) {
	sql := fmt.Sprintf("select "+CreateSelectString(&entities.HistoricalLocation{}, qo, "", "", hlMapping)+" FROM %s.historicallocation order by id desc "+CreateTopSkipQueryString(qo), gdb.Schema)
	countSql := fmt.Sprintf("select COUNT(*) FROM %s.historicallocation", gdb.Schema)
	return processHistoricalLocations(gdb.Db, sql, qo, countSql)
}

// GetHistoricalLocationsByLocation retrieves all historicallocations linked to the given location
func (gdb *GostDatabase) GetHistoricalLocationsByLocation(locationID interface{}, qo *odata.QueryOptions) ([]*entities.HistoricalLocation, int, error) {
	intID, ok := ToIntID(locationID)
	if !ok {
		return nil, 0, gostErrors.NewRequestNotFound(errors.New("Location does not exist"))
	}
	sql := fmt.Sprintf("select "+CreateSelectString(&entities.HistoricalLocation{}, qo, "", "", hlMapping)+" FROM %s.historicallocation where location_id = %v order by id desc "+CreateTopSkipQueryString(qo), gdb.Schema, intID)
	countSql := fmt.Sprintf("select COUNT(*) FROM %s.historicallocation where location_id = %v", gdb.Schema, intID)
	return processHistoricalLocations(gdb.Db, sql, qo, countSql)
}

// GetHistoricalLocationsByThing retrieves all historicallocations linked to the given thing
func (gdb *GostDatabase) GetHistoricalLocationsByThing(thingID interface{}, qo *odata.QueryOptions) ([]*entities.HistoricalLocation, int, error) {
	intID, ok := ToIntID(thingID)
	if !ok {
		return nil, 0, gostErrors.NewRequestNotFound(errors.New("Thing does not exist"))
	}
	sql := fmt.Sprintf("select "+CreateSelectString(&entities.HistoricalLocation{}, qo, "", "", hlMapping)+" FROM %s.historicallocation where thing_id = %v order by id desc "+CreateTopSkipQueryString(qo), gdb.Schema, intID)
	countSql := fmt.Sprintf("select COUNT(*) FROM %s.historicallocation where thing_id = %v", gdb.Schema, intID)
	return processHistoricalLocations(gdb.Db, sql, qo, countSql)
}

func processHistoricalLocation(db *sql.DB, sql string, qo *odata.QueryOptions) (*entities.HistoricalLocation, error) {
	hls, _, err := processHistoricalLocations(db, sql, qo, "")
	if err != nil {
		return nil, err
	}

	if len(hls) == 0 {
		return nil, gostErrors.NewRequestNotFound(errors.New("HistoricalLocation not found"))
	}

	return hls[0], nil
}

func processHistoricalLocations(db *sql.DB, sql string, qo *odata.QueryOptions, countSql string) ([]*entities.HistoricalLocation, int, error) {
	rows, err := db.Query(sql)
	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}

	var hls = []*entities.HistoricalLocation{}
	for rows.Next() {
		var id interface{}
		var time string

		var params []interface{}
		var qp []string
		if qo == nil || qo.QuerySelect == nil || len(qo.QuerySelect.Params) == 0 {
			s := &entities.HistoricalLocation{}
			qp = s.GetPropertyNames()
		} else {
			qp = qo.QuerySelect.Params
		}

		for _, p := range qp {
			if p == "id" {
				params = append(params, &id)
			}
			if p == "time" {
				params = append(params, &time)
			}
		}

		err = rows.Scan(params...)

		datastream := entities.HistoricalLocation{}
		datastream.ID = id
		datastream.Time = time

		hls = append(hls, &datastream)
	}

	var count int
	if len(countSql) > 0 {
		db.QueryRow(countSql).Scan(&count)
	}

	return hls, count, nil
}

// PostHistoricalLocation adds a historical location to the database
// returns the created historical location including the generated id
// fails when a thing or location cannot be found for the given id's
func (gdb *GostDatabase) PostHistoricalLocation(hl *entities.HistoricalLocation) (*entities.HistoricalLocation, error) {
	var hlID int
	tid, ok := ToIntID(hl.Thing.ID)
	if !ok || !gdb.ThingExists(tid) {
		return nil, gostErrors.NewRequestNotFound(errors.New("Thing does not exist"))
	}

	lid, ok := ToIntID(hl.Locations[0].ID)
	if !ok || !gdb.LocationExists(lid) {
		return nil, gostErrors.NewRequestNotFound(errors.New("Location does not exist"))
	}

	sql := fmt.Sprintf("INSERT INTO %s.historicallocation (time, thing_id, location_id) VALUES ($1, $2, $3) RETURNING id", gdb.Schema)
	err3 := gdb.Db.QueryRow(sql, time.Now(), tid, lid).Scan(&hlID)
	if err3 != nil {
		return nil, err3
	}

	hl.ID = hlID

	return hl, nil
}

// HistoricalLocationExists checks if a HistoricalLocation is present in the database based on a given id
func (gdb *GostDatabase) HistoricalLocationExists(locationID interface{}) bool {
	var result bool
	sql := fmt.Sprintf("SELECT exists (SELECT 1 FROM  %s.historicallocation WHERE id = $1 LIMIT 1)", gdb.Schema)
	err := gdb.Db.QueryRow(sql, locationID).Scan(&result)
	if err != nil {
		return false
	}

	return result
}

// PatchHistoricalLocation updates a HistoricalLocation in the database
func (gdb *GostDatabase) PatchHistoricalLocation(id interface{}, hl *entities.HistoricalLocation) (*entities.HistoricalLocation, error) {
	var err error
	var ok bool
	var intID int
	updates := make(map[string]interface{})

	if intID, ok = ToIntID(id); !ok || !gdb.HistoricalLocationExists(intID) {
		return nil, gostErrors.NewRequestNotFound(errors.New("HistoricalLocation does not exist"))
	}

	if len(hl.Time) > 0 {
		updates["time"] = hl.Time
	}

	if err = gdb.updateEntityColumns("historicallocation", updates, intID); err != nil {
		return nil, err
	}

	nhl, _ := gdb.GetHistoricalLocation(intID, nil)
	return nhl, nil
}

// DeleteHistoricalLocation tries to delete a HistoricalLocation by the given id
func (gdb *GostDatabase) DeleteHistoricalLocation(id interface{}) error {
	intID, ok := ToIntID(id)
	if !ok {
		return gostErrors.NewRequestNotFound(errors.New("HistoricalLocation does not exist"))
	}

	r, err := gdb.Db.Exec(fmt.Sprintf("DELETE FROM %s.historicallocation WHERE id = $1", gdb.Schema), intID)
	if err != nil {
		return err
	}

	if c, _ := r.RowsAffected(); c == 0 {
		return gostErrors.NewRequestNotFound(errors.New("HistoricalLocation not found"))
	}

	return nil
}
