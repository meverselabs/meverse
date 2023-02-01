package admin

const (
	AdminAdd               = "Admin.Add"
	AdminRemove            = "Admin.Remove"
	GeneratorAdd           = "Generator.Add"
	GeneratorRemove        = "Generator.Remove"
	ContractDeploy         = "Contract.Deploy"
	TransactionSetBasicFee = "Transaction.SetBasicFee"
)

func IsAdminMethod(method string) bool {
	if method == AdminAdd ||
		method == AdminRemove ||
		method == GeneratorAdd ||
		method == GeneratorRemove ||
		method == ContractDeploy ||
		method == TransactionSetBasicFee {
		return true
	}
	return false
}
