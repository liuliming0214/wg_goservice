package model

type Method struct {
	Method          string `json:"method"`
	Sys_app_secret  string `json:"sys_app_secret"`
	Sys_app_id      string `json:"sys_app_id"`
	Api_url         string `json:"api_url"`
	Is_error_record string `json:"is_error_record"`
	Limit_speed     uint64 `json:"limit_speed"`
	Is_layout       uint64 `json:"is_layout"`
}

func (Method) TableName() string {
	return "wg_system"
}
