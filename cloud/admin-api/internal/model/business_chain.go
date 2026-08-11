package model

// BusinessChain identifies which business chain a record belongs to.
type BusinessChain string

const (
	ChainSelf      BusinessChain = "self"
	ChainHospital  BusinessChain = "hospital"
	ChainCommunity BusinessChain = "community"
)

// SourceType identifies the origin of a medication rule.
type SourceType string

const (
	SourceCustom      SourceType = "custom"
	SourceDoctorOrder SourceType = "doctor_order"
	SourceCarePlan    SourceType = "care_plan"
)

// ChainPermissions maps role string → allowed business chains.
var ChainPermissions = map[string][]BusinessChain{
	"super_admin":     {ChainSelf, ChainHospital, ChainCommunity, "regulatory"},
	"operator":        {ChainSelf, "regulatory"},
	"hospital_doc":    {ChainHospital},
	"nurse":           {ChainHospital},
	"community_staff": {ChainCommunity},
	"regulator":       {ChainHospital, ChainCommunity, "regulatory"},
}
