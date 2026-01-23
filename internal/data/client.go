package data

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"stocker/internal/config"
	"stocker/pkg/models"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// SinaClient 新浪财经API客户端结构
type SinaClient struct {
	httpClient *http.Client
	config     *APIConfig
}

// NewSinaClient 创建新的新浪API客户端
func NewSinaClient(apiConfig *APIConfig) *SinaClient {
	return &SinaClient{
		httpClient: &http.Client{
			Timeout: apiConfig.Timeout,
		},
		config: apiConfig,
	}
}

// detectAndConvertEncoding 检测并转换编码
func detectAndConvertEncoding(data []byte) string {
	// 首先检查是否是有效的UTF-8
	if utf8.Valid(data) {
		return string(data)
	}

	// 尝试GBK解码（新浪财经常用GBK编码）
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		// 如果GBK解码也失败，使用普通UTF-8修复
		return strings.ToValidUTF8(string(data), strings.Repeat("?", 3))
	}

	// 检查转换后的结果是否是有效的UTF-8
	if utf8.Valid(result) {
		return string(result)
	}

	// 最后的fallback
	return strings.ToValidUTF8(string(result), strings.Repeat("?", 3))
}

// request 发送HTTP请求
func (c *SinaClient) request(url string, headers map[string]string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", config.SinaUserAgent)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 自动检测和转换编码
	result := detectAndConvertEncoding(body)

	return result, nil
}

// GetStockData 获取股票数据（主数据源，带重试机制）
func (c *SinaClient) GetStockData(symbol string) (*models.StockData, error) {
	var lastErr error

	// 使用主数据源重试
	for i := 0; i < c.config.RetryCount; i++ {
		if i > 0 {
			time.Sleep(c.config.RetryDelay)
		}

		data, err := c.getStockDataFromSina(symbol)
		if err == nil && data != nil {
			return data, nil
		}
		lastErr = err
	}

	// 主数据源失败，使用备用数据源
	data, err := c.getStockDataFromTencent(symbol)
	if err != nil && lastErr != nil {
		return nil, fmt.Errorf("主数据源和备用数据源均失败。主数据源错误: %v, 备用数据源错误: %v", lastErr, err)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getStockDataFromSina 从新浪API获取股票数据
func (c *SinaClient) getStockDataFromSina(symbol string) (*models.StockData, error) {
	url := config.SinaAPIURL + symbol
	headers := map[string]string{
		"Referer": config.SinaReferer,
	}

	response, err := c.request(url, headers)
	if err != nil {
		return nil, fmt.Errorf("新浪API请求失败: %w", err)
	}

	return parseSinaStockData(symbol, response)
}

// getStockDataFromTencent 从腾讯API获取股票数据
func (c *SinaClient) getStockDataFromTencent(symbol string) (*models.StockData, error) {
	// 转换代码格式，腾讯API的格式可能与新浪不同
	tencentSymbol := convertToTencentSymbol(symbol)

	url := config.TencentAPIURL + tencentSymbol

	response, err := c.request(url, nil)
	if err != nil {
		return nil, fmt.Errorf("腾讯API请求失败: %w", err)
	}

	return parseTencentStockData(symbol, response)
}

// BatchGetStockData 批量获取股票数据
func (c *SinaClient) BatchGetStockData(symbols []string) ([]*models.StockData, error) {
	var results []*models.StockData

	// 控制并发数量
	sem := make(chan struct{}, c.config.MaxConcurrent)

	errChan := make(chan error, len(symbols))
	dataChan := make(chan *models.StockData, len(symbols))

	for _, symbol := range symbols {
		go func(sym string) {
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			data, err := c.GetStockData(sym)
			if err != nil {
				errChan <- err
				return
			}
			dataChan <- data
		}(symbol)
	}

	// 等待所有goroutine完成
	for i := 0; i < len(symbols); i++ {
		select {
		case data := <-dataChan:
			results = append(results, data)
		case err := <-errChan:
			// 记录错误但继续处理其他股票
			fmt.Printf("警告: 股票数据获取失败: %v\n", err)
		}
	}

	return results, nil
}

// ValidateStock 验证股票代码是否有效
func (c *SinaClient) ValidateStock(symbol string) (bool, string) {
	data, err := c.GetStockData(symbol)
	if err != nil {
		return false, ""
	}
	return data != nil && data.Name != "", data.Name
}

// GetProviderName 获取提供商名称
func (c *SinaClient) GetProviderName() string {
	return "sina"
}

// convertToTencentSymbol 将新浪格式的股票代码转换为腾讯格式
func convertToTencentSymbol(symbol string) string {
	return symbol
}
