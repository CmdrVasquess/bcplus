package galaxy

import (
	"math"
	"os"
	"reflect"
	"testing"
)

func testRepo(t *testing.T) (res *Repo, filename string) {
	var err error
	if testing.Verbose() || true {
		filename = t.Name() + ".db"
		os.Remove(filename)
		res, err = NewRepo(filename)
	} else {
		filename = ""
		res, err = NewRepo(":memory:")
	}
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	err = runSqlFile(res.db, 0, "create-galaxy.sqlite.sql")
	if err != nil {
		t.Fatal(err)
	}
	return res, filename
}

func TestSystemCRUnD(t *testing.T) {
	rpo, _ := testRepo(t)
	defer rpo.Close()
	sys, err := rpo.PutSystem(&System{
		Name: "Achenar",
		Coos: Vec3D{1, 2, 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if sys.Id <= 0 {
		t.Errorf("invalid id: %d", sys.Id)
	}
	id := sys.Id
	sys, err = rpo.GetSystem(id)
	if err != nil {
		t.Fatal(err)
	}
	if sys == nil {
		t.Errorf("cannot find system by id %d", id)
	}
	if sys.Id != id {
		t.Errorf("system with id %d returned with id %d", id, sys.Id)
	}
	if sys.Name != "Achenar" {
		t.Errorf("unexpected system name '%s'", sys.Name)
	}
	if reflect.DeepEqual(sys.Coos, &Vec3D{1, 2, 3}) {
		t.Errorf("unexpected system coos %v", sys.Coos)
	}
	sys.Name = "Sol"
	sys.Coos[Xk] = -123
	_, err = rpo.PutSystem(sys)
	if err != nil {
		t.Fatal(err)
	}
	sys, err = rpo.FindSystem("Sol", nil)
	if err != nil {
		t.Fatal(err)
	}
	if sys == nil {
		t.Fatalf("cannot find system 'Sol'")
	}
	if sys.Name != "Sol" {
		t.Errorf("unexpected system name '%s'", sys.Name)
	}
	if sys.Coos[Xk] != -123 {
		t.Errorf("unexpected system coos %v", sys.Coos)
	}
}

func TestSysPartCRUnD(t *testing.T) {
	rpo, _ := testRepo(t)
	defer rpo.Close()
	sys, _ := rpo.PutSystem(&System{
		Name: "Sol",
		Coos: Vec3D{0, 0, 0},
	})
	_, err := sys.AddPart(&SysPart{
		Type:       Planet,
		Name:       "Earth",
		FromCenter: 1234,
		Lat:        float32(math.NaN()),
	})
	if err != nil {
		t.Fatal(err)
	}
	earthId := sys.Id
	sys, _ = rpo.FindSystem("Sol", nil)
	parts, err := sys.Parts()
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Errorf("unexpected number of parts: %d", len(parts))
	}
	earth := parts[0]
	if earth.Id != earthId {
		t.Errorf("wrong id for system part of sol: %d", earth.Id)
	}
	if earth.Name != "Earth" {
		t.Errorf("wrong id for system part of sol: %s", earth.Name)
	}
	earth.Height = -1
	_, err = rpo.PutSysPart(earth)
	if err != nil {
		t.Fatal(err)
	}
	earth, err = rpo.GetSysPart(earthId, nil)
	if err != nil {
		t.Fatal(err)
	}
	if earth.Height != -1 {
		t.Errorf("wrong updated height %f, expected -1", earth.Height)
	}
}

func TestResourceCnRUD(t *testing.T) {
	rpo, _ := testRepo(t)
	defer rpo.Close()
	sys, _ := rpo.PutSystem(&System{
		Name: "Sol",
		Coos: Vec3D{0, 0, 0},
	})
	loc, _ := sys.AddPart(&SysPart{
		Type:       Planet,
		Name:       "Earth",
		FromCenter: 1234,
		Lat:        float32(math.NaN()),
	})
	_, err := loc.AddResource(&Resource{
		Name:    "Beer",
		OccFreq: 0.67,
	})
	if err != nil {
		t.Fatal(err)
	}
}
