package base

import (
	"fmt"
	"strings"

	"zim.cn/base/log"

	"github.com/ttacon/libphonenumber"
)

var ErrInvalidPhoneNumber = fmt.Errorf("invalid phone number")

// 手机号归一化
func NormalizePhoneNumber(phone string) (string, error) {
	number := phone
	if !strings.HasPrefix(number, "+") {
		if !strings.HasPrefix(number, "86") {
			number = "86" + number
		}
		number = "+" + number
	}
	p, err := libphonenumber.Parse(number, "")
	if err != nil {
		return "", err
	}
	if !libphonenumber.IsValidNumber(p) {
		if p.GetCountryCode() == 86 {
			nationalNum := fmt.Sprintf("%d", *p.NationalNumber)
			if len(nationalNum) == 11 && (nationalNum[:3] == "199" || nationalNum[:3] == "166") {
				out := fmt.Sprintf("86%s", nationalNum)
				return out, nil
			}
		}
		return "", ErrInvalidPhoneNumber
	}

	out := libphonenumber.NormalizeDigitsOnly(number)
	return out, nil
}

// 手机号转地区代码
func PhoneToCC(phone_number string) string {
	if !strings.HasPrefix(phone_number, "+") {
		phone_number = "+" + phone_number
	}
	p, err := libphonenumber.Parse(phone_number, "")
	if err != nil {
		return ""
	}
	return libphonenumber.GetRegionCodeForNumber(p)
}

// 去除国家代码
func TrimPhoneCC(phone_number string, ccs ...string) string {
	for _, cc := range ccs {
		if len(cc) > 0 && strings.HasPrefix(phone_number, cc) {
			phone_number = strings.TrimPrefix(phone_number, cc)
		}
	}

	return phone_number
}

// 转化成ITU E.164格式, +CC MMMMMMMM
func ToE164(phone_number string) string {
	if phone_number == "" {
		return ""
	}
	if phone_number[:1] != "+" {
		phone_number = "+" + phone_number
	}
	p, err := libphonenumber.Parse(phone_number, "")
	if err != nil {
		log.Error("ToE164:", phone_number, err)
		return phone_number
	}
	return fmt.Sprintf("+%d %d", p.GetCountryCode(), p.GetNationalNumber())
}
