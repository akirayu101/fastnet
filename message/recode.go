package message

//请求的msgid

const SUCCESS int32 = 0
const (
	ERR_REGISTER int32 = 1 + iota
	ERR_LOGIN
	ERR_SELECTROLE
	ERR_BUYROLE
	ERR_LEVELUPROLE_NOT_EXISTS
	ERR_CASH_NOT_ENOUGH
	ERR_DIAMOND_NOT_ENOUGH
	ERR_SKILLLEVELUP_MAX_LEVEL
	ERR_ADDSKILL
	ERR_ADDSKILL_IS_EXISTS
	ERR_ADDSKILL_MAX
)
