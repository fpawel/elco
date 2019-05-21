package old

import (
	"database/sql"
	"github.com/fpawel/elco/internal/data"
	"math/rand"
	"strings"
	"time"
)

type Party struct {
	Date            OldPartyDate `json:"Date"`
	ID              string       `json:"Id"`
	Name            string       `json:"Name"`
	PGS1            float64      `json:"PGS1"`
	PGS2            float64      `json:"PGS2"`
	PGS3            float64      `json:"PGS3"`
	ProductType     string       `json:"ProductType"`
	Products        []Product    `json:"Products"`
	ProductsSerials []int64      `json:"ProductsSerials"`
}

type Product struct {
	ID               string     `json:"Id"`
	N                int        `json:"N"`
	Serial           int64      `json:"Serial"`
	IsReportIncluded bool       `json:"IsReportIncluded"`
	Flash            []byte     `json:"Flash"`
	ProductType      string     `json:"ProductType"`
	IsChecked        bool       `json:"IsChecked"`
	IsCustomTermo    bool       `json:"IsCustomTermo"`
	Ifon             *float64   `json:"Ifon"`
	Isns             *float64   `json:"Isns"`
	IfMinus20        *float64   `json:"If_20"`
	IsMinus20        *float64   `json:"Is_20"`
	If50             *float64   `json:"If50"`
	Is50             *float64   `json:"Is50"`
	I13              *float64   `json:"I13"`
	I24              *float64   `json:"I24"`
	I35              *float64   `json:"I35"`
	I26              *float64   `json:"I26"`
	I17              *float64   `json:"I17"`
	In               *float64   `json:"In"`
	CustomTermo      []struct{} `json:"CustomTermo"`
}

type OldPartyDate struct {
	Day         int        `json:"Day"`
	Hour        int        `json:"Hour"`
	Millisecond int        `json:"Millisecond"`
	Minute      int        `json:"Minute"`
	Month       time.Month `json:"Month"`
	Second      int        `json:"Second"`
	Year        int        `json:"Year"`
}

func (x Party) Party() (p data.Party, products []data.Product) {
	p.Concentration1 = x.PGS1
	p.Concentration2 = x.PGS2
	p.Concentration3 = x.PGS3
	p.PointsMethod = 3
	p.ProductTypeName = x.ProductType
	p.Note.String = x.Name
	p.Note.Valid = len(strings.TrimSpace(x.Name)) > 0

	f := func(a *float64) sql.NullFloat64 {
		if a == nil {
			return sql.NullFloat64{}
		}
		return sql.NullFloat64{
			Valid:   true,
			Float64: *a,
		}
	}

	for _, y := range x.Products {
		if y.Serial == 0 {
			continue
		}
		products = append(products, data.Product{
			Serial: sql.NullInt64{
				Int64: y.Serial,
				Valid: true,
			},
			Place:       y.N,
			Production:  y.IsReportIncluded,
			Firmware:    y.Flash,
			I13:         f(y.I13),
			I24:         f(y.I24),
			I35:         f(y.I35),
			I26:         f(y.I26),
			I17:         f(y.I17),
			IFPlus20:    f(y.Ifon),
			IFMinus20:   f(y.IfMinus20),
			IFPlus50:    f(y.If50),
			ISPlus20:    f(y.Isns),
			ISMinus20:   f(y.IsMinus20),
			ISPlus50:    f(y.Is50),
			NotMeasured: f(y.In),
		})
	}
	return
}

func NewOldParty(s data.Party, products []data.Product) (p Party) {

	t := time.Now()
	p = Party{
		ID:          randStringBytesMaskImprSrc(12),
		PGS1:        s.Concentration1,
		PGS2:        s.Concentration2,
		PGS3:        s.Concentration3,
		ProductType: s.ProductTypeName,
		Name:        s.Note.String,
		Date: OldPartyDate{
			Year:   t.Year(),
			Month:  t.Month(),
			Day:    t.Day(),
			Hour:   t.Hour(),
			Minute: t.Minute(),
			Second: t.Second(),
		},
		Products:        make([]Product, 64),
		ProductsSerials: make([]int64, 64),
	}

	f := func(a sql.NullFloat64) *float64 {
		if !a.Valid {
			return nil
		}
		v := a.Float64
		return &v
	}

	for i, y := range products {

		p.Products[y.Place].N = i
		if y.Place >= 64 {
			continue
		}
		p.ProductsSerials[y.Place] = y.Serial.Int64
		p.Products[y.Place] = Product{
			ID:               randStringBytesMaskImprSrc(12),
			Serial:           y.Serial.Int64,
			N:                y.Place,
			IsReportIncluded: y.Production,
			IsChecked:        true,
			I13:              f(y.I13),
			I24:              f(y.I24),
			I35:              f(y.I35),
			I26:              f(y.I26),
			I17:              f(y.I17),
			Ifon:             f(y.IFPlus20),
			IfMinus20:        f(y.IFMinus20),
			If50:             f(y.IFPlus50),
			Isns:             f(y.ISPlus20),
			IsMinus20:        f(y.ISMinus20),
			Is50:             f(y.ISPlus50),
			In:               f(y.NotMeasured),
			CustomTermo:      []struct{}{},
			ProductType:      y.ProductTypeName.String,
		}
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
