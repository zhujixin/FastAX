// 安全护栏模型
//
// 本文件包含安全护栏相关的数据模型：
//   - GuardrailRule: 护栏规则表，定义 PII 检测、注入检测、敏感词过滤等规则
//   - GuardrailLog: 护栏检测日志表，记录每次检测的结果和处理方式
package model

// GuardrailRule 护栏规则表，定义 PII 检测、注入检测、敏感词过滤等规则
type GuardrailRule struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"size:128;not null"`
	Stage      string `gorm:"size:16;not null"`   // before | after
	Type       string `gorm:"size:32;not null"`    // pii | injection | secret | content
	Action     string `gorm:"size:16;not null"`    // enforce | monitor | log
	Conditions string `gorm:"type:text"`
	Priority   int    `gorm:"default:0"`
	Enabled    int    `gorm:"default:1"`
	CreatedAt  int64  `gorm:"autoCreateTime"`
}

type GuardrailLog struct {
	ID               uint   `gorm:"primaryKey"`
	TraceID          string `gorm:"index;size:64;not null"`
	UserID           uint   `gorm:"index"`
	RuleID           uint   `gorm:"index"`
	Stage            string `gorm:"size:16;not null"`
	DetectedEntities string `gorm:"type:text"`
	ActionTaken      string `gorm:"size:16;not null"`
	CreatedAt        int64  `gorm:"autoCreateTime;index"`
}
