package usecases

import (
	"io"
	"project_masAde/internal/repository"
)

type DashboardUsecase struct {
	configRepo   *repository.ConfigRepository
	tableManager *repository.TableManager
}

func NewDashboardUsecase(configRepo *repository.ConfigRepository, tableManager *repository.TableManager) *DashboardUsecase {
	return &DashboardUsecase{
		configRepo:   configRepo,
		tableManager: tableManager,
	}
}

// Product Management - DEPRECATED: Use dynamic datasets instead
// GetAllProducts returns empty - migrate to dynamic tables
func (u *DashboardUsecase) GetAllProducts() []repository.Product {
	return []repository.Product{}
}

// Config Management (tenant-aware)
func (u *DashboardUsecase) GetConfig(schemaName, key string) (string, error) {
	return u.configRepo.GetConfig(schemaName, key)
}

func (u *DashboardUsecase) SetConfig(schemaName, key, value string) error {
	return u.configRepo.SetConfig(schemaName, key, value)
}

func (u *DashboardUsecase) GetAllConfigs(schemaName string) ([]repository.BotConfig, error) {
	return u.configRepo.GetAllConfigs(schemaName)
}

// Menu Management (tenant-aware)
func (u *DashboardUsecase) GetMenu(schemaName, slug string) (*repository.Menu, error) {
	return u.configRepo.GetMenu(schemaName, slug)
}

func (u *DashboardUsecase) CreateMenu(schemaName string, m *repository.Menu) error {
	return u.configRepo.CreateMenu(schemaName, m)
}

func (u *DashboardUsecase) UpdateMenu(schemaName string, m *repository.Menu) error {
	return u.configRepo.UpdateMenu(schemaName, m)
}

func (u *DashboardUsecase) DeleteMenu(schemaName, slug string) error {
	return u.configRepo.DeleteMenu(schemaName, slug)
}

func (u *DashboardUsecase) GetAllMenus(schemaName string) ([]repository.Menu, error) {
	return u.configRepo.GetAllMenus(schemaName)
}

// Dynamic Data Management (tenant-aware)
func (u *DashboardUsecase) ImportTable(schemaName, displayName string, csvData io.Reader) error {
	return u.tableManager.ImportCSV(schemaName, displayName, csvData)
}

func (u *DashboardUsecase) ListTables(schemaName string) ([]repository.TableMetadata, error) {
	return u.tableManager.ListTables(schemaName)
}

func (u *DashboardUsecase) GetTableData(schemaName, tableName string) ([]map[string]interface{}, error) {
	return u.tableManager.GetTableData(schemaName, tableName)
}

func (u *DashboardUsecase) DeleteTable(schemaName, tableName string) error {
	return u.tableManager.DeleteTable(schemaName, tableName)
}

func (u *DashboardUsecase) UpdateRow(schemaName, tableName string, rowID int, data map[string]interface{}) error {
	return u.tableManager.UpdateRow(schemaName, tableName, rowID, data)
}

func (u *DashboardUsecase) DeleteRow(schemaName, tableName string, rowID int) error {
	return u.tableManager.DeleteRow(schemaName, tableName, rowID)
}
