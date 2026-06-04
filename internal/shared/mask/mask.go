// 数据脱敏工具
//
// 本文件实现敏感数据的脱敏函数：
//   - MaskPhone: 手机号中间4位脱敏（138****8000）
//   - MaskEmail: 邮箱@前部分脱敏（j***@example.com）
//   - MaskIDNumber: 身份证号首尾各留4位
//   - MaskBankCard: 银行卡号仅显示后4位
//   - MaskAPIKey: API Key 首尾各4字符
package mask

import "strings"

// MaskPhone 手机号脱敏：保留前3后4，中间用****替代
// 13800138000 → 138****8000
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return strings.Repeat("*", len(phone))
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskEmail 邮箱脱敏：@前仅显示首字符，其余用***替代
// john.doe@example.com → j***@example.com
func MaskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 1 {
		return email
	}
	return email[:1] + "***" + email[at:]
}

// MaskIDNumber 身份证号脱敏：保留前4后4，中间用****替代
func MaskIDNumber(id string) string {
	if len(id) < 8 {
		return strings.Repeat("*", len(id))
	}
	return id[:4] + strings.Repeat("*", len(id)-8) + id[len(id)-4:]
}

// MaskBankCard 银行卡号脱敏：仅显示后4位
func MaskBankCard(card string) string {
	if len(card) < 4 {
		return strings.Repeat("*", len(card))
	}
	return "**** **** **** " + card[len(card)-4:]
}

// MaskAPIKey API Key 脱敏：显示前2和后2字符（≥4字符时），更短则完全遮蔽
func MaskAPIKey(key string) string {
	if len(key) < 4 {
		return strings.Repeat("*", len(key))
	}
	if len(key) <= 6 {
		return key[:2] + strings.Repeat("*", len(key)-4) + key[len(key)-2:]
	}
	// For longer keys, show first 4 and last 4
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
