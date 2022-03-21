package generator

import (
	"errors"

	"github.com/flosch/pongo2/v5"
)

var AZURE_REGION_DISPLAY_MAP_SHORT = map[string]string{
	"eastasia":           "ea",
	"southeastasia":      "sea",
	"centralus":          "cus",
	"eastus":             "eus",
	"eastus2":            "eus2",
	"westus":             "wus",
	"northcentralus":     "ncus",
	"southcentralus":     "scus",
	"northeurope":        "ne",
	"westeurope":         "we",
	"japanwest":          "jw",
	"japaneast":          "je",
	"brazilsouth":        "bs",
	"australiaeast":      "ae",
	"australiasoutheast": "ase",
	"southindia":         "si",
	"centralindia":       "ci",
	"westindia":          "wi",
	"canadacentral":      "cc",
	"canadaeast":         "ce",
	"uksouth":            "us",
	"ukwest":             "uw",
	"westcentralus":      "wc",
	"westus2":            "wus2",
	"koreacentral":       "kc",
	"koreasouth":         "ks",
	"francecentral":      "fc",
	"francesouth":        "fs",
	"australiacentral":   "ac",
	"australiacentral2":  "ac2",
	"southafricanorth":   "san",
	"southafricawest":    "saw",
}

var AZURE_REGION_DISPLAY_MAP_LONG = map[string]string{
	"eastasia":           "East Asia",
	"southeastasia":      "Southeast Asia",
	"centralus":          "Central US",
	"eastus":             "East US",
	"eastus2":            "East US 2",
	"westus":             "West US",
	"northcentralus":     "North Central US",
	"southcentralus":     "South Central US",
	"northeurope":        "North Europe",
	"westeurope":         "West Europe",
	"japanwest":          "Japan West",
	"japaneast":          "Japan East",
	"brazilsouth":        "Brazil South",
	"australiaeast":      "Australia East",
	"australiasoutheast": "Australia Southeast",
	"southindia":         "South India",
	"centralindia":       "Central India",
	"westindia":          "West India",
	"canadacentral":      "Canada Central",
	"canadaeast":         "Canada East",
	"uksouth":            "UK South",
	"ukwest":             "UK West",
	"westcentralus":      "West Central US",
	"westus2":            "West US 2",
	"koreacentral":       "Korea Central",
	"koreasouth":         "Korea South",
	"francecentral":      "France Central",
	"francesouth":        "France South",
	"australiacentral":   "Australia Central",
	"australiacentral2":  "Australia Central 2",
	"southafricanorth":   "South Africa North",
	"southafricawest":    "South Africa West",
}

func filterAzureRegionShort(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	key := in.String()
	if val, ok := AZURE_REGION_DISPLAY_MAP_SHORT[key]; ok {
		return pongo2.AsValue(val), nil
	}
	return nil, &pongo2.Error{
		Sender:    "filter:azure_region_short",
		OrigError: errors.New("region not found"),
	}
}

func filterAzureRegionLong(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	key := in.String()
	if val, ok := AZURE_REGION_DISPLAY_MAP_LONG[key]; ok {
		return pongo2.AsValue(val), nil
	}
	return nil, &pongo2.Error{
		Sender:    "filter:azure_region_long",
		OrigError: errors.New("region not found"),
	}
}
