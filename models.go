package hjem

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrInvalidZipCodes     = errors.New("invalid zip code format")
	ErrInvalidPropertyType = errors.New("invalid property type format")
	ErrUnknownAddr         = errors.New("unknown address")
)

type Config struct {
	PropertyTypes []string
	ZipCodeFrom   *int
	ZipCodeTo     *int
}

func NewConfig(zipcodes, properties string) (*Config, error) {
	var conf Config

	if zipcodes != "" {
		zip, err := strconv.Atoi(zipcodes)
		if err == nil {
			return &Config{
				ZipCodeFrom: &zip,
				ZipCodeTo:   &zip,
			}, nil
		}

		splits := strings.Split(zipcodes, "-")
		if len(splits) != 2 {
			return nil, ErrInvalidZipCodes
		}

		from, err := strconv.Atoi(splits[0])
		if err != nil {
			return nil, ErrInvalidZipCodes
		}

		to, err := strconv.Atoi(splits[1])
		if err != nil {
			return nil, ErrInvalidZipCodes
		}

		conf.ZipCodeFrom = &from
		conf.ZipCodeTo = &to
	}

	if properties != "" {
		valid := map[string]bool{
			"house":     true,
			"apartment": true,
			"vacation":  true,
		}

		splits := strings.Split(properties, ",")
		if len(splits) == 0 {
			if _, ok := valid[properties]; !ok {
				return nil, ErrInvalidPropertyType
			}
			conf.PropertyTypes = []string{properties}
		} else {
			for _, p := range strings.Split(properties, ",") {
				if _, ok := valid[p]; !ok {
					return nil, ErrInvalidPropertyType
				}
			}

			conf.PropertyTypes = splits
		}
	}

	return &conf, nil
}

// type Address struct {
// 	ID               uint           `gorm:"primaryKey"`
// 	DawaID           string         `gorm:"not null,unique"`
// 	StreetName       string         `gorm:"index:addr_idx,not null"`
// 	StreetNumber     string         `gorm:"index:addr_idx,not null"`
// 	Floor            sql.NullString `gorm:"index:addr_idx"`
// 	Door             sql.NullString `gorm:"index:addr_idx"`
// 	PostalCode       int            `gorm:"index:addr_idx,not null"`
// 	MunicipalityCode string         `gorm:"not null"`
// 	Longtitude       float64        `gorm:"not null"`
// 	Latitude         float64        `gorm:"not null"`
// }

// type Sale struct {
// 	AddressID uint      `gorm:"primaryKey"`
// 	SoldDate  time.Time `gorm:"primaryKey"`

// 	BoligaEstateId int `gorm:"not null"`
// 	AmountDKK      int `gorm:"not null"`
// 	PropertyType   int `gorm:"not null"`
// 	SqMeters       int `gorm:"not null"`
// 	Rooms          int
// 	BuildYear      int `gorm:"not null"`
// 	PriceChange    float64
// 	SaleType       string `gorm:"not null"`
// }

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) (*Store, error) {
	db.AutoMigrate(&Address{})
	// db.AutoMigrate(&Sale{})

	return &Store{
		db: db,
	}, nil
}

func (s *Store) CountAddresses() int64 {
	var count int64
	s.db.Model(Address{}).Count(&count)
	return count
}

func (s *Store) StreamAddrs(addrC <-chan Address) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var batch []Address
		for addr := range addrC {
			if len(batch) > 100 {
				s.db.Create(&batch)
				batch = []Address{}
			}

			batch = append(batch, addr)
		}

		if len(batch) > 0 {
			s.db.Create(&batch)
			batch = []Address{}
		}

		return nil
	})
}

var (
	addrRegexp = regexp.MustCompile(`([a-zA-Z\p{L}-. ]+) ([0-9][a-zA-Z0-9\p{L}]*)(, ([a-zA-Z0-9\p{L}]+)\.?)?( ([a-zA-Z0-9\p{L}]+)\.?)?`)
)

// func (s *Store) SaveSale(sale BoligaSale) error {
// 	matches := addrRegexp.FindAllStringSubmatch(sale.Addr, 1)
// 	if len(matches) == 0 {
// 		return fmt.Errorf("unable to parse address")
// 	}

// 	query := map[string]interface{}{
// 		"street_name":   matches[0][1],
// 		"street_number": matches[0][2],
// 		"floor":         nil,
// 		"door":          nil,
// 		"postal_code":   sale.ZipCode,
// 	}

// 	if floor := matches[0][4]; floor != "" {
// 		query["floor"] = floor
// 	}

// 	if door := matches[0][6]; door != "" {
// 		query["door"] = door
// 	}

// 	var addrs []Address
// 	s.db.Find(&addrs, query)

// 	if len(addrs) == 0 {
// 		results, _ := DawaFuzzySearch{fmt.Sprintf("%s, %d", sale.Addr, sale.ZipCode)}.Fetch()
// 		if len(results) != 1 {
// 			return ErrUnknownAddr
// 		}

// 		s.db.Find(&addrs, map[string]interface{}{"dawa_id": results[0].ID})
// 		if len(addrs) != 1 {
// 			return ErrUnknownAddr
// 		}
// 	}

// 	addr := addrs[0]
// 	saleR := Sale{
// 		AddressID:      addr.ID,
// 		SoldDate:       sale.SoldDate,
// 		BoligaEstateId: sale.EstateId,
// 		AmountDKK:      sale.AmountDKK,
// 		PropertyType:   sale.PropertyType,
// 		SqMeters:       sale.SqMeters,
// 		Rooms:          int(sale.Rooms),
// 		BuildYear:      sale.BuildYear,
// 		PriceChange:    sale.PriceChange,
// 		SaleType:       sale.SaleType,
// 	}

// 	s.db.Create(&saleR)
// 	return nil
// }
