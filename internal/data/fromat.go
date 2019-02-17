package data

import (
	"fmt"
	"gopkg.in/reform.v1"
	"strings"
)

func (s Product) String2() string {
	res := make([]string, 23)
	res[0] = "ProductID: " + reform.Inspect(s.ProductID, false)
	res[1] = "PartyID: " + reform.Inspect(s.PartyID, false)
	res[2] = "Serial: " + reform.Inspect(s.Serial, false)
	res[3] = "Place: " + reform.Inspect(s.Place, false)
	res[4] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, false)
	res[5] = "Note: " + reform.Inspect(s.Note, true)
	res[6] = "IFMinus20: " + reform.Inspect(s.IFMinus20, false)
	res[7] = "IFPlus20: " + reform.Inspect(s.IFPlus20, false)
	res[8] = "IFPlus50: " + reform.Inspect(s.IFPlus50, false)
	res[9] = "ISMinus20: " + reform.Inspect(s.ISMinus20, false)
	res[10] = "ISPlus20: " + reform.Inspect(s.ISPlus20, false)
	res[11] = "ISPlus50: " + reform.Inspect(s.ISPlus50, false)
	res[12] = "I13: " + reform.Inspect(s.I13, false)
	res[13] = "I24: " + reform.Inspect(s.I24, false)
	res[14] = "I35: " + reform.Inspect(s.I35, false)
	res[15] = "I26: " + reform.Inspect(s.I26, false)
	res[16] = "I17: " + reform.Inspect(s.I17, false)
	res[17] = "NotMeasured: " + reform.Inspect(s.NotMeasured, false)
	res[18] = "Firmware: " + reform.Inspect(len(s.Firmware), false) + " байт"
	res[19] = "Production: " + reform.Inspect(s.Production, false)
	res[20] = "OldProductID: " + reform.Inspect(s.OldProductID, false)
	res[21] = "OldSerial: " + reform.Inspect(s.OldSerial, false)
	res[22] = "PointsMethod: " + reform.Inspect(s.PointsMethod, false)
	return strings.Join(res, ", ")
}

func (s ProductInfo) String2() string {
	res := make([]string, 49)
	res[0] = "ProductID: " + reform.Inspect(s.ProductID, false)
	res[1] = "PartyID: " + reform.Inspect(s.PartyID, false)
	res[2] = "Serial: " + reform.Inspect(s.Serial, false)
	res[3] = "Place: " + reform.Inspect(s.Place, false)
	res[4] = "CreatedAt: " + reform.Inspect(s.CreatedAt, false)
	res[5] = "IFMinus20: " + reform.Inspect(s.IFMinus20, false)
	res[6] = "IFPlus20: " + reform.Inspect(s.IFPlus20, false)
	res[7] = "IFPlus50: " + reform.Inspect(s.IFPlus50, false)
	res[8] = "ISMinus20: " + reform.Inspect(s.ISMinus20, false)
	res[9] = "ISPlus20: " + reform.Inspect(s.ISPlus20, false)
	res[10] = "ISPlus50: " + reform.Inspect(s.ISPlus50, false)
	res[11] = "I13: " + reform.Inspect(s.I13, false)
	res[12] = "I24: " + reform.Inspect(s.I24, false)
	res[13] = "I35: " + reform.Inspect(s.I35, false)
	res[14] = "I26: " + reform.Inspect(s.I26, false)
	res[15] = "I17: " + reform.Inspect(s.I17, false)
	res[16] = "NotMeasured: " + reform.Inspect(s.NotMeasured, false)
	res[17] = "KSensMinus20: " + reform.Inspect(s.KSensMinus20, false)
	res[18] = "KSens20: " + reform.Inspect(s.KSens20, false)
	res[19] = "KSens50: " + reform.Inspect(s.KSens50, false)
	res[20] = "DFon20: " + reform.Inspect(s.DFon20, false)
	res[21] = "DFon50: " + reform.Inspect(s.DFon50, false)
	res[22] = "DNotMeasured: " + reform.Inspect(s.DNotMeasured, false)
	res[23] = "OKMinFon20: " + reform.Inspect(s.OKMinFon20, false)
	res[24] = "OKMaxFon20: " + reform.Inspect(s.OKMaxFon20, false)
	res[25] = "OKMinFon20r: " + reform.Inspect(s.OKMinFon20r, false)
	res[26] = "OKMaxFon20r: " + reform.Inspect(s.OKMaxFon20r, false)
	res[27] = "OKDFon20: " + reform.Inspect(s.OKDFon20, false)
	res[28] = "OKMinKSens20: " + reform.Inspect(s.OKMinKSens20, false)
	res[29] = "OKMaxKSens20: " + reform.Inspect(s.OKMaxKSens20, false)
	res[30] = "OKMinKSens50: " + reform.Inspect(s.OKMinKSens50, false)
	res[31] = "OKMaxKSens50: " + reform.Inspect(s.OKMaxKSens50, false)
	res[32] = "OKDFon50: " + reform.Inspect(s.OKDFon50, false)
	res[33] = "OKDNotMeasured: " + reform.Inspect(s.OKDNotMeasured, false)
	res[34] = "Ok: " + reform.Inspect(s.Ok, false)
	res[35] = "HasFirmware: " + reform.Inspect(s.HasFirmware, false)
	res[36] = "Production: " + reform.Inspect(s.Production, false)
	res[37] = "AppliedProductTypeName: " + reform.Inspect(s.AppliedProductTypeName, false)
	res[38] = "GasCode: " + reform.Inspect(s.GasCode, false)
	res[39] = "UnitsCode: " + reform.Inspect(s.UnitsCode, false)
	res[40] = "GasName: " + reform.Inspect(s.GasName, false)
	res[41] = "UnitsName: " + reform.Inspect(s.UnitsName, false)
	res[42] = "Scale: " + reform.Inspect(s.Scale, false)
	res[43] = "NobleMetalContent: " + reform.Inspect(s.NobleMetalContent, false)
	res[44] = "LifetimeMonths: " + reform.Inspect(s.LifetimeMonths, false)
	res[45] = "PointsMethod: " + reform.Inspect(s.PointsMethod, false)
	res[46] = "AppliedPointsMethod: " + reform.Inspect(s.AppliedPointsMethod, false)
	res[47] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, false)
	res[48] = "Note: " + reform.Inspect(s.Note, false)
	return strings.Join(res, ", ")
}

func (s PartyInfo) String2() string {
	res := make([]string, 4)
	res[0] = "PartyID: " + reform.Inspect(s.PartyID, false)
	res[1] = "CreatedAt: " + reform.Inspect(s.CreatedAt, false)
	res[2] = "UpdatedAt: " + reform.Inspect(s.UpdatedAt, false)
	res[3] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, false)
	return strings.Join(res, ", ")
}

func (s Party) String2() string {
	res := make([]string, 4)
	res[0] = "PartyID: " + reform.Inspect(s.PartyID, false)
	res[1] = "CreatedAt: " + reform.Inspect(s.CreatedAt, false)
	res[2] = "UpdatedAt: " + reform.Inspect(s.UpdatedAt, false)
	res[3] = "ProductTypeName: " + reform.Inspect(s.ProductTypeName, false)
	return strings.Join(res, ", ")
}

func (x Firmware) String2() string {
	return strings.Join([]string{
		"Place: " + reform.Inspect(x.Place, false),
		"CreatedAt: " + reform.Inspect(x.CreatedAt, false),
		"Serial: " + reform.Inspect(x.Serial, false),
		"Units: " + reform.Inspect(x.Units, false),
		"Gas: " + reform.Inspect(x.Gas, false),
		"KSens20: " + reform.Inspect(x.KSens20, false),
		"ScaleBegin: " + reform.Inspect(x.ScaleBegin, false),
		"ScaleEnd: " + reform.Inspect(x.ScaleEnd, false),
		"Fon: " + reform.Inspect(x.Fon, false),
		"Sens: " + reform.Inspect(x.Sens, false),
		fmt.Sprintf("Bytes: % X", x.Bytes()),
	}, ", ")
}
