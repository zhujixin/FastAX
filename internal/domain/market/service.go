package market

import (
	"github.com/fastax/fastax-server/internal/shared/model"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type ModelInfo struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
	Price    string `json:"price"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func (s *Service) ListModels(provider, modelType string) ([]ModelInfo, error) {
	// Query from platform token products
	var products []model.TokenProduct
	query := s.db.Where("status = 1").Order("sort_order asc")
	if modelType != "" {
		query = query.Where("type = ?", modelType)
	}
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	// Get supplier info
	supplierIDs := make(map[uint]bool)
	for _, p := range products {
		supplierIDs[p.SupplierID] = true
	}
	var suppliers []model.Supplier
	s.db.Find(&suppliers)
	supplierMap := make(map[uint]model.Supplier)
	for _, sp := range suppliers {
		supplierMap[sp.ID] = sp
	}

	var result []ModelInfo
	for _, p := range products {
		sp := supplierMap[p.SupplierID]
		if provider != "" && sp.Code != provider {
			continue
		}
		result = append(result, ModelInfo{
			ID:       p.ID,
			Name:     p.Name,
			Model:    p.Model,
			Provider: sp.Name,
			Price:    p.Price,
			Currency: p.Currency,
			Status:   "available",
		})
	}

	// Also include vendor products
	var vendorProducts []model.SupplierProduct
	vQuery := s.db.Where("status = 'active'")
	if err := vQuery.Find(&vendorProducts).Error; err == nil {
		vendorIDs := make(map[uint]bool)
		for _, vp := range vendorProducts {
			vendorIDs[vp.VendorID] = true
		}
		var vendors []model.SupplierVendor
		s.db.Find(&vendors)
		vendorMap := make(map[uint]model.SupplierVendor)
		for _, v := range vendors {
			vendorMap[v.ID] = v
		}

		for _, vp := range vendorProducts {
			v := vendorMap[vp.VendorID]
			if provider != "" && v.CompanyName != provider {
				continue
			}
			result = append(result, ModelInfo{
				ID:       vp.ID,
				Name:     vp.Name,
				Model:    vp.Model,
				Provider: v.CompanyName,
				Price:    vp.Price,
				Currency: vp.Currency,
				Status:   "available",
			})
		}
	}

	return result, nil
}

// --- Compare & Health ---

type ModelComparison struct {
	Model          string                `json:"model"`
	Providers      []ProviderModelDetail `json:"providers"`
	ContextWindow  int                   `json:"context_window"`
}

type ProviderModelDetail struct {
	ProviderID   uint    `json:"provider_id"`
	ProviderName string  `json:"provider_name"`
	Price        string  `json:"price"`
	Currency     string  `json:"currency"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
	Status       string  `json:"status"`
}

func (s *Service) CompareModels(modelNames []string) ([]ModelComparison, error) {
	if len(modelNames) == 0 {
		return []ModelComparison{}, nil
	}

	// Query all active token products matching the model names
	var products []model.TokenProduct
	if err := s.db.Where("model IN ? AND status = 1", modelNames).Find(&products).Error; err != nil {
		return nil, err
	}

	// Build supplier map
	supplierIDs := make(map[uint]bool)
	for _, p := range products {
		supplierIDs[p.SupplierID] = true
	}
	var suppliers []model.Supplier
	s.db.Find(&suppliers)
	supplierMap := make(map[uint]model.Supplier)
	for _, sp := range suppliers {
		supplierMap[sp.ID] = sp
	}

	// Build health map (latest per provider)
	healthMap := make(map[uint]model.ProviderHealth)
	var allHealth []model.ProviderHealth
	s.db.Order("period_end desc").Find(&allHealth)
	for _, h := range allHealth {
		if _, exists := healthMap[h.ProviderID]; !exists {
			healthMap[h.ProviderID] = h
		}
	}

	// Group by model
	modelMap := make(map[string][]ProviderModelDetail)
	for _, p := range products {
		sp := supplierMap[p.SupplierID]
		health := healthMap[p.SupplierID]
		status := "healthy"
		if health.Status == 0 {
			status = "unknown"
		}
		detail := ProviderModelDetail{
			ProviderID:   p.SupplierID,
			ProviderName: sp.Name,
			Price:        p.Price,
			Currency:     p.Currency,
			AvgLatencyMs: health.AvgLatencyMs,
			ErrorRate:    health.ErrorRate,
			Status:       status,
		}
		modelMap[p.Model] = append(modelMap[p.Model], detail)
	}

	var result []ModelComparison
	for _, name := range modelNames {
		providers, ok := modelMap[name]
		if !ok {
			providers = []ProviderModelDetail{}
		}
		result = append(result, ModelComparison{
			Model:         name,
			Providers:     providers,
			ContextWindow: 0,
		})
	}
	return result, nil
}

type ProviderHealthRecord struct {
	ID           uint    `json:"id"`
	ProviderID   uint    `json:"provider_id"`
	ProviderName string  `json:"provider_name"`
	Status       int     `json:"status"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
	CheckCount   int     `json:"check_count"`
	PeriodStart  int64   `json:"period_start"`
	PeriodEnd    int64   `json:"period_end"`
}

func (s *Service) GetProviderHealth(providerID uint) ([]ProviderHealthRecord, error) {
	var records []model.ProviderHealth
	if err := s.db.Where("provider_id = ?", providerID).Order("period_end desc").Find(&records).Error; err != nil {
		return nil, err
	}
	var supplier model.Supplier
	s.db.First(&supplier, providerID)
	result := make([]ProviderHealthRecord, len(records))
	for i, r := range records {
		result[i] = ProviderHealthRecord{
			ID:           r.ID,
			ProviderID:   r.ProviderID,
			ProviderName: supplier.Name,
			Status:       r.Status,
			AvgLatencyMs: r.AvgLatencyMs,
			ErrorRate:    r.ErrorRate,
			CheckCount:   r.CheckCount,
			PeriodStart:  r.PeriodStart,
			PeriodEnd:    r.PeriodEnd,
		}
	}
	return result, nil
}

type ProviderStatus struct {
	ProviderID   uint    `json:"provider_id"`
	ProviderName string  `json:"provider_name"`
	Status       int     `json:"status"`
	AvgLatencyMs int     `json:"avg_latency_ms"`
	ErrorRate    float64 `json:"error_rate"`
}

func (s *Service) ListProviders() ([]ProviderStatus, error) {
	var suppliers []model.Supplier
	if err := s.db.Find(&suppliers).Error; err != nil {
		return nil, err
	}
	// Latest health per provider
	healthMap := make(map[uint]model.ProviderHealth)
	var allHealth []model.ProviderHealth
	s.db.Order("period_end desc").Find(&allHealth)
	for _, h := range allHealth {
		if _, exists := healthMap[h.ProviderID]; !exists {
			healthMap[h.ProviderID] = h
		}
	}
	var result []ProviderStatus
	for _, sp := range suppliers {
		health := healthMap[sp.ID]
		status := 0
		avgLatency := 0
		var errRate float64
		if health.ID != 0 {
			status = health.Status
			avgLatency = health.AvgLatencyMs
			errRate = health.ErrorRate
		}
		result = append(result, ProviderStatus{
			ProviderID:   sp.ID,
			ProviderName: sp.Name,
			Status:       status,
			AvgLatencyMs: avgLatency,
			ErrorRate:    errRate,
		})
	}
	return result, nil
}

// --- Model Variants ---

type VariantRequest struct {
	BaseModel       string  `json:"base_model" binding:"required"`
	Suffix          string  `json:"suffix" binding:"required"`
	ProviderID      uint    `json:"provider_id" binding:"required"`
	PriceCoefficient float64 `json:"price_coefficient"`
	Priority        int     `json:"priority"`
}

func (s *Service) CreateVariant(req *VariantRequest) (*model.ModelVariant, error) {
	coef := req.PriceCoefficient
	if coef == 0 {
		coef = 1.0
	}
	variant := model.ModelVariant{
		BaseModel:       req.BaseModel,
		Suffix:          req.Suffix,
		ProviderID:      req.ProviderID,
		PriceCoefficient: coef,
		Priority:        req.Priority,
	}
	if err := s.db.Create(&variant).Error; err != nil {
		return nil, err
	}
	return &variant, nil
}

func (s *Service) ListVariants(baseModel string) ([]model.ModelVariant, error) {
	var variants []model.ModelVariant
	query := s.db.Order("priority desc")
	if baseModel != "" {
		query = query.Where("base_model = ?", baseModel)
	}
	if err := query.Find(&variants).Error; err != nil {
		return nil, err
	}
	return variants, nil
}

func (s *Service) FindVariant(modelName string) (*model.ModelVariant, error) {
	var variants []model.ModelVariant
	if err := s.db.Find(&variants).Error; err != nil {
		return nil, err
	}
	for _, v := range variants {
		if modelName == v.BaseModel+v.Suffix {
			return &v, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
