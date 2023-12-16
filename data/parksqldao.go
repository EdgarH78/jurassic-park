package data

import (
	"database/sql"
	"fmt"

	"github.com/EdgarH78/jurassic-park/models"
	_ "github.com/go-sql-driver/mysql"
)

var (
	maxOpenConns = 20
	maxIdleConns = 10
)

type SQLConfig struct {
	User         string
	Password     string
	Host         string
	DatabaseName string
}

func (sc *SQLConfig) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", sc.User, sc.Password, sc.Host, sc.DatabaseName)
}

type ParkSqlDao struct {
	db *sql.DB
}

func NewParkSqlDao(sqlConfig SQLConfig) (*ParkSqlDao, error) {
	db, err := sql.Open("mysql", sqlConfig.ConnectionString())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &ParkSqlDao{
		db: db,
	}, nil
}

func (s *ParkSqlDao) AddCage(cage models.Cage) error {
	qs := `INSERT INTO cage(externalId, capacity, hasPower)
			VALUES(?,?,?)`
	params := []interface{}{cage.Label, cage.MaxOccupancy, cage.HasPower}
	_, err := s.db.Exec(qs, params...)
	if err != nil {
		return err
	}
	return nil
}

func (s *ParkSqlDao) GetCage(cageLabel string) (*models.Cage, error) {
	cage, _, err := s.getCageWithId(cageLabel)
	if err != nil {
		return nil, err
	}
	return cage, nil
}

func (s *ParkSqlDao) getCageWithId(cageLabel string) (*models.Cage, int, error) {
	qs := `SELECT id, externalId, capacity, hasPower 
			FROM cage			
			WHERE externalId = ?
			`
	rows, err := s.db.Query(qs, cageLabel)
	defer rows.Close()

	if err != nil {
		return nil, 0, err
	}
	if !rows.Next() {
		return nil, 0, models.EntityNotFound
	}
	var id int
	cage := models.Cage{}
	if err := rows.Scan(&id, &cage.Label, &cage.MaxOccupancy, &cage.HasPower); err != nil {
		return nil, 0, err
	}

	occupancy, err := s.getDinosaurCountInCage(id)
	if err != nil {
		return nil, 0, err
	}
	cage.Occupancy = occupancy

	return &cage, id, nil
}

func (s *ParkSqlDao) GetCages() ([]models.Cage, error) {
	qs := `SELECT id, externalId, capacity, hasPower 
			FROM cage			
			ORDER BY id`
	rows, err := s.db.Query(qs)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	cages := []models.Cage{}
	for rows.Next() {
		var id int
		cage := models.Cage{}
		if err := rows.Scan(&id, &cage.Label, &cage.MaxOccupancy, &cage.HasPower); err != nil {
			return nil, err
		}
		occupancy, err := s.getDinosaurCountInCage(id)
		if err != nil {
			return nil, err
		}
		cage.Occupancy = occupancy
		cages = append(cages, cage)
	}

	return cages, nil
}

func (s *ParkSqlDao) getDinosaurCountInCage(cageId int) (int, error) {
	qs := `SELECT COUNT(*) FROM dinosaur where cageId=?`
	rows, err := s.db.Query(qs, cageId)
	defer rows.Close()

	if err != nil {
		return 0, err
	}
	if !rows.Next() {
		// This shouldn't happen, but if it does the the cage must have been deleted
		return 0, models.EntityNotFound
	}
	var occupancy int
	if err := rows.Scan(&occupancy); err != nil {
		return 0, err
	}
	return occupancy, nil
}

func (s *ParkSqlDao) AddDinosaur(dinosaur models.Dinosaur) error {
	speciesQuery := `SELECT COUNT(*) FROM species where name=?`
	speciesRows, err := s.db.Query(speciesQuery, dinosaur.Species)
	defer speciesRows.Close()

	if err != nil {
		return err
	}
	if !speciesRows.Next() {
		// This shouldn't happen, so treat it like an internal server error
		return fmt.Errorf("no data returned from the database when checking for species %s", dinosaur.Species)
	}

	var speciesCount int
	err = speciesRows.Scan(&speciesCount)
	if err != nil {
		return err
	}
	if speciesCount == 0 {
		return models.InvalidDinosaurSpecies
	}

	insertStmt := `INSERT IGNORE INTO dinosaur(name, species, sex)
					VALUES(?,?,'Female')`
	params := []interface{}{dinosaur.Name, dinosaur.Species}
	result, err := s.db.Exec(insertStmt, params...)
	if err != nil {
		return err
	}

	rowsInserted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsInserted == 0 {
		// no rows were added, and that means there is already a dinosaur with that name
		return models.EntityAlreadyExists
	}

	return nil
}

func (s *ParkSqlDao) GetDinosaurs() ([]models.Dinosaur, error) {
	qs := `SELECT d.name, d.species, s.diet, c.externalId
		   FROM dinosaur d
		   JOIN species s on s.name=d.species 
		   LEFT OUTER JOIN cage c on c.id=d.cageId
		   ORDER BY d.id`
	rows, err := s.db.Query(qs)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	dinosaurs := []models.Dinosaur{}
	for rows.Next() {
		dinosaur := models.Dinosaur{}
		err = rows.Scan(&dinosaur.Name, &dinosaur.Species, &dinosaur.Diet, &dinosaur.Cage)
		if err != nil {
			return nil, err
		}
		dinosaurs = append(dinosaurs, dinosaur)
	}
	return dinosaurs, nil
}

func (s *ParkSqlDao) GetDinosaur(name string) (*models.Dinosaur, error) {
	qs := `SELECT d.name, d.species, s.diet, c.externalId
		   FROM dinosaur d
		   JOIN species s on s.name=d.species 
		   LEFT OUTER JOIN cage c on c.id=d.cageId
		   WHERE d.name=?`
	rows, err := s.db.Query(qs, name)
	defer rows.Close()

	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, models.EntityNotFound
	}

	dinosaur := models.Dinosaur{}
	err = rows.Scan(&dinosaur.Name, &dinosaur.Species, &dinosaur.Diet, &dinosaur.Cage)
	if err != nil {
		return nil, err
	}
	return &dinosaur, nil
}

func (s *ParkSqlDao) AddDinosaurToCage(dinosaurName, targetCage string) error {
	dinosaur, err := s.GetDinosaur(dinosaurName)
	if err != nil {
		return err
	}
	cage, cageId, err := s.getCageWithId(targetCage)
	if err != nil {
		return err
	}
	if cage.Occupancy >= cage.MaxOccupancy {
		return models.CageCapacityExceeded
	}
	if !cage.HasPower {
		return models.IncompatibleCagePowerState
	}
	if dinosaur.Diet == "Carnivore" {
		cageHasOtherSpecies, err := s.cageHasOtherSpecies(*dinosaur, *cage)
		if err != nil {
			return err
		}
		if cageHasOtherSpecies {
			return models.IncompatibleSpecies
		}
	} else {
		cageHasCarnivores, err := s.cageHasCarnivores(*cage)
		if err != nil {
			return err
		}
		if cageHasCarnivores {
			return models.IncompatibleSpecies
		}
	}

	updateStatement := `UPDATE dinosaur 
		   SET cageId=?
		   WHERE name=?`
	params := []interface{}{cageId, dinosaur.Name}
	_, err = s.db.Exec(updateStatement, params...)
	if err != nil {
		return err
	}
	return nil
}

func (s *ParkSqlDao) cageHasOtherSpecies(dinosaur models.Dinosaur, cage models.Cage) (bool, error) {
	qs := `SELECT COUNT(*)
		   FROM cage c 
		   JOIN dinosaur d on d.cageId=c.id 
		   WHERE c.externalId=? AND d.species<>?`
	rows, err := s.db.Query(qs, cage.Label, dinosaur.Species)
	defer rows.Close()

	if err != nil {
		return false, err
	}
	if !rows.Next() {
		return false, nil
	}
	var otherSpeciesCount int
	err = rows.Scan(&otherSpeciesCount)
	if err != nil {
		return false, err
	}

	return otherSpeciesCount > 0, nil
}

func (s *ParkSqlDao) cageHasCarnivores(cage models.Cage) (bool, error) {
	qs := `SELECT COUNT(*)
		   FROM cage c 
		   JOIN dinosaur d on d.cageId=c.id
		   JOIN species s on s.name=d.species
		   WHERE c.externalId=? AND s.diet='Carnivore'`
	rows, err := s.db.Query(qs, cage.Label)
	defer rows.Close()

	if err != nil {
		return false, err
	}
	if !rows.Next() {
		return false, nil
	}
	var carnivoreCount int
	err = rows.Scan(&carnivoreCount)
	if err != nil {
		return false, err
	}

	return carnivoreCount > 0, nil
}

func (s *ParkSqlDao) GetDinosaursInCage(cageLabel string) ([]models.Dinosaur, error) {
	cage, cageId, err := s.getCageWithId(cageLabel)
	if err != nil {
		return nil, err
	}

	qs := `SELECT d.name, d.species, s.diet
		   FROM dinosaur d
		   JOIN species s on s.name=d.species 
		   WHERE d.cageId=?
		   ORDER BY d.id`
	rows, err := s.db.Query(qs, cageId)
	defer rows.Close()

	dinosaurs := []models.Dinosaur{}
	for rows.Next() {
		dinosaur := models.Dinosaur{
			Cage: &cage.Label,
		}
		err = rows.Scan(&dinosaur.Name, &dinosaur.Species, &dinosaur.Diet)
		if err != nil {
			return nil, err
		}
		dinosaurs = append(dinosaurs, dinosaur)
	}
	return dinosaurs, nil
}

func (s *ParkSqlDao) UpdateCagePowerStatus(cageLabel string, powerOn bool) error {
	cage, cageId, err := s.getCageWithId(cageLabel)
	if err != nil {
		return err
	}
	if !powerOn && cage.Occupancy > 0 {
		return models.IncompatibleCagePowerState
	}

	updateStatement := `UPDATE cage
				   SET hasPower=?
				   WHERE id=?`
	params := []interface{}{powerOn, cageId}
	_, err = s.db.Exec(updateStatement, params...)
	if err != nil {
		return err
	}
	return nil

}
