//  This file is part of the Eliona project.
//  Copyright © 2024 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package dbhelper

import (
	appmodel "app-name/app/model"
	dbgen "app-name/db/generated"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

// DBHelper is a singleton struct managing the database connection and queries.
type DBHelper struct {
	db      *sql.DB
	queries *dbgen.Queries
}

var (
	instance *DBHelper
)

// InitDB initializes the database connection ONCE.
func InitDB(db *sql.DB) {
	instance = &DBHelper{
		db:      db,
		queries: dbgen.New(db), // Create sqlc Queries instance
	}
}

// GetDB returns the singleton database instance.
func GetDB() *DBHelper {
	if instance == nil {
		log.Fatal("Database not initialized. Call InitDB() first.")
	}
	return instance
}

// GetQueries returns the sqlc Queries instance.
func GetQueries() *dbgen.Queries {
	if instance == nil {
		log.Fatal("Database not initialized. Call InitDB() first.")
	}
	return instance.queries
}

// CloseDB gracefully shuts down the database connection.
func CloseDB() error {
	if instance != nil && instance.db != nil {
		return instance.db.Close()
	}
	return nil
}

// Insert a new configuration using sqlc-generated queries
func InsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config: %v", err)
	}
	queries := GetQueries()
	insertedConfig, err := queries.InsertConfig(ctx,
		dbgen.InsertConfigParams{
			ApiAccessChangeMe: dbConfig.ApiAccessChangeMe,
			RefreshInterval:   dbConfig.RefreshInterval,
			RequestTimeout:    dbConfig.RequestTimeout,
			AssetFilter:       dbConfig.AssetFilter,
			Active:            dbConfig.Active,
			Enable:            dbConfig.Enable,
			ProjectIds:        dbConfig.ProjectIds,
			UserID:            dbConfig.UserID,
		})
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}

	return toAppConfig(insertedConfig)
}

// Upsert (insert or update) a configuration
func UpsertConfig(ctx context.Context, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config: %v", err)
	}

	queries := GetQueries()

	upsertedConfig, err := queries.UpsertConfig(ctx,
		dbgen.UpsertConfigParams{
			ID:                dbConfig.ID,
			ApiAccessChangeMe: dbConfig.ApiAccessChangeMe,
			RefreshInterval:   dbConfig.RefreshInterval,
			RequestTimeout:    dbConfig.RequestTimeout,
			AssetFilter:       dbConfig.AssetFilter,
			Active:            dbConfig.Active,
			Enable:            dbConfig.Enable,
			ProjectIds:        dbConfig.ProjectIds,
			UserID:            dbConfig.UserID,
		})
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("upserting DB config: %v", err)
	}

	return toAppConfig(upsertedConfig)
}

// Fetch a configuration by ID
func GetConfig(ctx context.Context, configID int64) (appmodel.Configuration, error) {
	queries := GetQueries()
	dbConfig, err := queries.GetConfig(ctx, configID)
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Configuration{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("fetching config: %v", err)
	}
	return toAppConfig(dbConfig)
}

// Delete a configuration
func DeleteConfig(ctx context.Context, configID int64) error {
	queries := GetQueries()
	err := queries.DeleteConfig(ctx, configID)
	if err != nil {
		return fmt.Errorf("deleting config: %v", err)
	}
	return nil
}

// Convert an application configuration to a database configuration for insertion/upsert
func toDbConfig(ctx context.Context, appConfig appmodel.Configuration) (dbgen.AppSchemaNameConfiguration, error) {
	af, err := json.Marshal(appConfig.AssetFilter)
	if err != nil {
		return dbgen.AppSchemaNameConfiguration{}, fmt.Errorf("marshalling assetFilter: %v", err)
	}

	dbConfig := dbgen.AppSchemaNameConfiguration{
		ID:                appConfig.Id,
		ApiAccessChangeMe: appConfig.ApiAccessChangeMe,
		RefreshInterval:   appConfig.RefreshInterval,
		RequestTimeout:    appConfig.RequestTimeout,
		AssetFilter:       af,
		Active:            appConfig.Active,
		Enable:            appConfig.Enable,
		ProjectIds:        appConfig.ProjectIDs,
	}

	env := frontend.GetEnvironment(ctx)
	if env != nil {
		dbConfig.UserID = env.UserId
	}

	return dbConfig, nil
}

// Convert a database configuration to an application model
func toAppConfig(dbConfig dbgen.AppSchemaNameConfiguration) (appmodel.Configuration, error) {
	var af [][]appmodel.FilterRule
	if err := json.Unmarshal(dbConfig.AssetFilter, &af); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("unmarshalling assetFilter: %v", err)
	}

	return appmodel.Configuration{
		ApiAccessChangeMe: dbConfig.ApiAccessChangeMe,
		Id:                dbConfig.ID,
		Enable:            dbConfig.Enable,
		RefreshInterval:   dbConfig.RefreshInterval,
		RequestTimeout:    dbConfig.RequestTimeout,
		AssetFilter:       af,
		Active:            dbConfig.Active,
		ProjectIDs:        dbConfig.ProjectIds,
		UserId:            dbConfig.UserID,
	}, nil
}

// Fetch all configurations
func GetConfigs(ctx context.Context) ([]appmodel.Configuration, error) {
	queries := GetQueries()
	dbConfigs, err := queries.GetConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching configs: %v", err)
	}

	var appConfigs []appmodel.Configuration
	for _, dbConfig := range dbConfigs {
		ac, err := toAppConfig(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("converting config: %v", err)
		}
		appConfigs = append(appConfigs, ac)
	}
	return appConfigs, nil
}

// Update a specific configuration's active state
func SetConfigActiveState(ctx context.Context, config appmodel.Configuration, state bool) error {
	queries := GetQueries()
	return queries.SetConfigActiveState(ctx, dbgen.SetConfigActiveStateParams{
		ID:     config.Id,
		Active: state,
	})
}

// Set all configurations inactive
func SetAllConfigsInactive(ctx context.Context) error {
	queries := GetQueries()
	return queries.SetAllConfigsInactive(ctx)
}

// Insert a new asset into the database
func InsertAsset(ctx context.Context, config appmodel.Configuration, projId string, globalAssetID string, assetId int32, providerId string) error {
	queries := GetQueries()
	return queries.InsertAsset(ctx, dbgen.InsertAssetParams{
		ConfigurationID: config.Id,
		ProjectID:       projId,
		GlobalAssetID:   globalAssetID,
		ProviderID:      providerId,
		AssetID:         sql.NullInt32{Int32: assetId, Valid: true},
	})
}

// Retrieve an asset ID by configuration, project, and global asset ID
func GetAssetId(ctx context.Context, config appmodel.Configuration, projId string, globalAssetID string) (*int32, error) {
	queries := GetQueries()
	assetID, err := queries.GetAssetId(ctx, dbgen.GetAssetIdParams{
		ConfigurationID: config.Id,
		ProjectID:       projId,
		GlobalAssetID:   globalAssetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if assetID.Valid {
		return nil, nil
	}
	return &assetID.Int32, nil
}

// Fetch an asset by ID including its related configuration
func GetAssetById(assetId int32) (appmodel.Asset, error) {
	queries := GetQueries()
	row, err := queries.GetAssetById(context.Background(), int64(assetId))
	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}

	return toAppAssetFromRow(row)
}

// Convert a database asset to an application model asset
func toAppAsset(dbAsset dbgen.AppSchemaNameAsset, config appmodel.Configuration) appmodel.Asset {
	return appmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		ProjectID:     dbAsset.ProjectID,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
	}
}

func toAppAssetFromRow(row dbgen.GetAssetByIdRow) (appmodel.Asset, error) {
	// Convert the AssetFilter field (JSON) into the expected [][]appmodel.FilterRule type
	var assetFilter [][]appmodel.FilterRule
	if err := json.Unmarshal(row.AssetFilter, &assetFilter); err != nil {
		return appmodel.Asset{}, fmt.Errorf("unmarshalling asset filter: %v", err)
	}

	// Construct the Configuration object
	config := appmodel.Configuration{
		ApiAccessChangeMe: row.ApiAccessChangeMe,
		Id:                row.ID_2, // Configuration ID
		Enable:            row.Enable,
		RefreshInterval:   row.RefreshInterval,
		RequestTimeout:    row.RequestTimeout,
		AssetFilter:       assetFilter,
		Active:            row.Active,
		ProjectIDs:        row.ProjectIds,
		UserId:            row.UserID,
	}

	// Construct the Asset object
	asset := appmodel.Asset{
		ID:            row.ID, // Asset ID
		Config:        config, // Configuration struct
		ProjectID:     row.ProjectID,
		GlobalAssetID: row.GlobalAssetID,
		ProviderID:    row.ProviderID,
		AssetID:       row.AssetID.Int32, // Extracting value from sql.NullInt32
	}

	return asset, nil
}
