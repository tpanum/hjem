package hjem

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	addrEndpoint = "https://api.dataforsyningen.dk/adresser"
)

var (
	ErrNotFoundAddr = errors.New("Unable to find address")
)

type DAWAAddress struct {
	FullText         string  `json:"betegnelse""`
	StreetName       string  `json:"vejnavn"`
	StreetNumber     string  `json:"husnr"`
	Floor            *string `json:"etage"`
	Door             *string `json:"dør"`
	PostalCode       string  `json:"postnr"`
	MunicipalityCode string  `json:"kommunekode"`
	Latitude         float64 `json:"x"`
	Longtitude       float64 `json:"y"`
}

type Address struct {
	ID               uint    `json:"-" gorm:"primaryKey"`
	DawaID           string  `json:"full_txt" gorm:"not null;unique"`
	StreetName       string  `json:"street_name" gorm:"not null"`
	StreetNumber     string  `json:"street_number" gorm:"not null"`
	Floor            *string `json:"floor"`
	Door             *string `json:"door"`
	PostalCode       string  `json:"zipcode" gorm:"not null"`
	MunicipalityCode string  `json:"municipality_code" gorm:"not null"`
	Latitude         float64 `json:"lat" gorm:"not null"`
	Longtitude       float64 `json:"long" gorm:"not null"`

	BoligaCollectedAt         time.Time    `json:"-"`
	BoligaPropertyKind        PropertyType `json:"-"`
	BoligaBuildingSize        int          `json:"building_size"`
	BoligaPropertySize        int          `json:"property_size"`
	BoligaBasementSize        int          `json:"basement_size"`
	BoligaRooms               int          `json:"rooms"`
	BoligaBuiltYear           int          `json:"built_year"`
	BoligaMonthlyOwnerExpense int          `json:"monthly_owner_expense_dkk"`
	BoligaEnergyMarking       string       `json:"energy_marking"`
}

func (addr Address) Short() string {
	s := fmt.Sprintf("%s %s", addr.StreetName, addr.StreetNumber)
	if addr.Floor != nil {
		s += fmt.Sprintf(", %s", *addr.Floor)
	}

	if addr.Door != nil {
		s += fmt.Sprintf(". %s", *addr.Door)
	}

	return s
}

func (a Address) ToSlice() []string {
	var door string
	if a.Door != nil {
		door = *a.Door
	}
	var floor string
	if a.Floor != nil {
		floor = *a.Floor
	}

	return []string{
		a.DawaID,
		a.StreetName,
		a.StreetNumber,
		door,
		floor,
		a.PostalCode,
		strconv.Itoa(a.BoligaBuildingSize),
		strconv.Itoa(a.BoligaPropertySize),
		strconv.Itoa(a.BoligaBasementSize),
		strconv.Itoa(a.BoligaRooms),
		strconv.Itoa(a.BoligaBuiltYear),
		strconv.Itoa(a.BoligaMonthlyOwnerExpense),
	}
}

func (a Address) Headers() []string {
	return []string{
		"id",
		"street_name",
		"street_number",
		"door",
		"floor",
		"postal_code",
		"size",
		"property_size",
		"basement_size",
		"rooms",
		"built_year",
		"monthly_owner_expense_dkk",
	}
}

type dawaCacher struct {
	db        *gorm.DB
	maxAmount float64
}

type DawaQueryCache struct {
	Query     string `gorm:"not null,unique"`
	IDs       string `gorm:"not null"`
	CreatedAt time.Time
}

func NewDawaQueryCacheFromAddrs(req DawaRequest, addrs []*Address) DawaQueryCache {
	reqStr := fmt.Sprintf("%s", req.Request().URL)

	ids := make([]string, len(addrs))
	for i := 0; i < len(addrs); i++ {
		ids[i] = strconv.Itoa(int(addrs[i].ID))
	}

	return DawaQueryCache{
		Query: reqStr,
		IDs:   strings.Join(ids, ","),
	}
}

func (dqc DawaQueryCache) Identifiers() []int {
	ids := strings.Split(dqc.IDs, ",")

	uids := make([]int, len(ids))
	for i := 0; i < len(ids); i++ {
		id, _ := strconv.Atoi(ids[i])
		uids[i] = id
	}

	return uids
}

type DawaCacher interface {
	Do(DawaRequest) ([]*Address, error)
}

func NewDawaCacher(db *gorm.DB) *dawaCacher {
	db.AutoMigrate(&DawaQueryCache{})
	db.AutoMigrate(&Address{})

	return &dawaCacher{
		maxAmount: 50.0,
		db:        db,
	}
}

func (c dawaCacher) Do(req DawaRequest) ([]*Address, error) {
	reqStr := fmt.Sprintf("%s", req.Request().URL)

	var cache DawaQueryCache
	var performRequest bool
	if err := c.db.First(&cache, "query = ?", reqStr).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		performRequest = true
	}

	if time.Now().Sub(cache.CreatedAt) >= req.MaxAge() {
		performRequest = true
	}

	if performRequest {
		if cache.Query != "" {
			if err := c.db.Delete(&cache).Error; err != nil {
				return nil, err
			}
		}

		addrs, err := req.Fetch()
		if err := c.safeCreateOrGetAddrs(addrs); err != nil {
			return nil, err
		}

		cache := NewDawaQueryCacheFromAddrs(req, addrs)
		if err := c.db.Create(&cache).Error; err != nil {
			return nil, err
		}

		return addrs, err
	}

	var addrs []*Address
	ids := cache.Identifiers()
	r := int(math.Ceil(float64(len(ids)) / c.maxAmount))
	for i := 0; i < r; i++ {
		var tempAddrs []*Address
		start, end := int(c.maxAmount)*i, int(c.maxAmount)*(i+1)
		end = int(math.Min(float64(len(ids)), float64(end)))
		if err := c.db.Find(&tempAddrs, ids[start:end]).Error; err != nil {
			return nil, err
		}

		addrs = append(addrs, tempAddrs...)
	}

	return addrs, nil
}

func (c dawaCacher) safeCreateOrGetAddrs(addrs []*Address) error {
	n := float64(len(addrs))
	r := int(math.Ceil(n / c.maxAmount))

	m := map[string]*Address{}
	for i := 0; i < r; i++ {
		start, end := int(c.maxAmount)*i, int(c.maxAmount)*(i+1)
		end = int(math.Min(n, float64(end)))

		var tempAddrs []*Address
		ids := make([]string, end-start)
		for j, a := range addrs[start:end] {
			ids[j] = a.DawaID
		}

		if err := c.db.Where("dawa_id IN ?", ids).Find(&tempAddrs).Error; err != nil {
			return err
		}

		for j, _ := range tempAddrs {
			a := tempAddrs[j]
			m[a.DawaID] = a
		}
	}

	var createAddrs []*Address
	for i, _ := range addrs {
		a := addrs[i]
		exsts, ok := m[a.DawaID]
		if !ok {
			createAddrs = append(createAddrs, a)
			continue
		}

		addrs[i] = exsts
	}

	if err := c.db.CreateInBatches(&createAddrs, int(c.maxAmount)).Error; err != nil {
		return err
	}

	return nil
}

type DawaRequest interface {
	Request() *http.Request
	MaxAge() time.Duration
	Fetch() ([]*Address, error)
}

func reqToAddrs(req *http.Request) ([]*Address, error) {
	q := req.URL.Query()
	q.Add("struktur", "mini")
	req.URL.RawQuery = q.Encode()

	resp, err := DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var temp []*DAWAAddress
	if err := json.NewDecoder(resp.Body).Decode(&temp); err != nil {
		return nil, err
	}

	output := make([]*Address, len(temp))
	for i, d := range temp {
		output[i] = &Address{
			DawaID:           d.FullText,
			StreetName:       d.StreetName,
			StreetNumber:     d.StreetNumber,
			Floor:            d.Floor,
			Door:             d.Door,
			PostalCode:       d.PostalCode,
			MunicipalityCode: d.MunicipalityCode,
			Latitude:         d.Latitude,
			Longtitude:       d.Longtitude,
		}
	}

	return output, nil
}

type DawaFuzzySearch struct {
	Query string
}

func (dfs DawaFuzzySearch) Request() *http.Request {
	req, err := http.NewRequest("GET", addrEndpoint, nil)
	if err != nil {
		return nil
	}

	q := req.URL.Query()
	q.Add("q", dfs.Query)
	req.URL.RawQuery = q.Encode()

	return req
}

func (dfs DawaFuzzySearch) Fetch() ([]*Address, error) {
	req := dfs.Request()
	return reqToAddrs(req)
}

func (dfs DawaFuzzySearch) MaxAge() time.Duration {
	return 365 * 24 * time.Hour
}

type DawaNearbySearch struct {
	Addr   Address
	Meters int
}

func (dns DawaNearbySearch) Request() *http.Request {
	req, _ := http.NewRequest("GET", addrEndpoint, nil)
	qStr := fmt.Sprintf("%f,%f,%d", dns.Addr.Latitude, dns.Addr.Longtitude, dns.Meters)

	q := req.URL.Query()
	q.Add("cirkel", qStr)
	req.URL.RawQuery = q.Encode()

	return req
}

func (dns DawaNearbySearch) Fetch() ([]*Address, error) {
	req := dns.Request()
	return reqToAddrs(req)
}

func (dfs DawaNearbySearch) MaxAge() time.Duration {
	return 365 * 24 * time.Hour
}
