package data

import "stocker/pkg/models"

// ProviderChain 提供主备提供商封装（主失败才切换）。
type ProviderChain struct {
	primary  StockProvider
	fallback StockProvider
}

func NewProviderChain(primary, fallback StockProvider) StockProvider {
	if fallback == nil {
		return primary
	}
	return &ProviderChain{primary: primary, fallback: fallback}
}

func (c *ProviderChain) GetStockData(symbol string) (*models.StockData, error) {
	data, err := c.primary.GetStockData(symbol)
	if err == nil && data != nil {
		return data, nil
	}
	if c.fallback == nil {
		return data, err
	}
	return c.fallback.GetStockData(symbol)
}

func (c *ProviderChain) BatchGetStockData(symbols []string) ([]*models.StockData, error) {
	data, err := c.primary.BatchGetStockData(symbols)
	if err == nil && len(data) > 0 {
		return data, nil
	}
	if c.fallback == nil {
		return data, err
	}
	return c.fallback.BatchGetStockData(symbols)
}

func (c *ProviderChain) SearchStock(keyword string) ([]models.SearchStock, error) {
	data, err := c.primary.SearchStock(keyword)
	if err == nil && len(data) > 0 {
		return data, nil
	}
	if c.fallback == nil {
		return data, err
	}
	return c.fallback.SearchStock(keyword)
}

func (c *ProviderChain) ValidateStock(symbol string) (bool, string) {
	ok, name := c.primary.ValidateStock(symbol)
	if ok && name != "" {
		return ok, name
	}
	if c.fallback == nil {
		return ok, name
	}
	return c.fallback.ValidateStock(symbol)
}

func (c *ProviderChain) GetProviderName() string {
	if c.fallback == nil {
		return c.primary.GetProviderName()
	}
	return c.primary.GetProviderName() + "+" + c.fallback.GetProviderName()
}
