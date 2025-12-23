package usecases

import (
	"io"
	"project_masAde/internal/repository"
)

type DashboardUsecase struct {
	configRepo   *repository.ConfigRepository
	productRepo  *repository.ProductRepository
	tableManager *repository.TableManager
}

func NewDashboardUsecase(configRepo *repository.ConfigRepository, productRepo *repository.ProductRepository, tableManager *repository.TableManager) *DashboardUsecase {
	return &DashboardUsecase{
		configRepo:   configRepo,
		productRepo:  productRepo,
		tableManager: tableManager,
	}
}

// Product Management
func (u *DashboardUsecase) GetAllProducts() []repository.Product {
	return u.productRepo.GetAllProducts()
}

// Config Management
func (u *DashboardUsecase) GetConfig(key string) (string, error) {
	return u.configRepo.GetConfig(key)
}

func (u *DashboardUsecase) SetConfig(key, value string) error {
	return u.configRepo.SetConfig(key, value)
}

func (u *DashboardUsecase) GetAllConfigs() ([]repository.BotConfig, error) {
	return u.configRepo.GetAllConfigs()
}

// Menu Management
func (u *DashboardUsecase) GetMenu(slug string) (*repository.Menu, error) {
	return u.configRepo.GetMenu(slug)
}

func (u *DashboardUsecase) CreateMenu(m *repository.Menu) error {
	return u.configRepo.CreateMenu(m)
}

func (u *DashboardUsecase) UpdateMenu(m *repository.Menu) error {
	return u.configRepo.UpdateMenu(m)
}

func (u *DashboardUsecase) DeleteMenu(slug string) error {
	return u.configRepo.DeleteMenu(slug)
}

func (u *DashboardUsecase) GetAllMenus() ([]repository.Menu, error) {
	return u.configRepo.GetAllMenus()
}

// Dynamic Data Management
func (u *DashboardUsecase) ImportTable(displayName string, csvData io.Reader) error {
	return u.tableManager.ImportCSV(displayName, csvData)
}

func (u *DashboardUsecase) ListTables() ([]repository.TableMetadata, error) {
	return u.tableManager.ListTables()
}

func (u *DashboardUsecase) GetTableData(tableName string) ([]map[string]interface{}, error) {
	return u.tableManager.GetTableData(tableName)
}

func (u *DashboardUsecase) DeleteTable(tableName string) error {
	return u.tableManager.DeleteTable(tableName)
}

func (u *DashboardUsecase) UpdateRow(tableName string, rowID int, data map[string]interface{}) error {
	return u.tableManager.UpdateRow(tableName, rowID, data)
}

func (u *DashboardUsecase) DeleteRow(tableName string, rowID int) error {
	return u.tableManager.DeleteRow(tableName, rowID)
}
