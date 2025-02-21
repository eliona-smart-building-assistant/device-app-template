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
	dbgen "app-name/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/stephenafamo/bob"
	null "github.com/volatiletech/null/v8"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

func InsertConfig(ctx context.Context, db bob.Executor, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}

	insertID, err := dbgen.Configurations.Insert(&dbConfig).Exec(ctx, db)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}

	return config, nil
}

func UpsertConfig(ctx context.Context, db bob.Executor, config appmodel.Configuration) (appmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}

	err = dbgen.InsertConfiguration(dbConfig).
		OnConflict(dbgen.ConfigurationColumns.ID).DoUpdate().
		Exec(ctx, db)

	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("upserting DB config: %v", err)
	}

	return config, nil
}

func GetConfig(ctx context.Context, db bob.Executor, configID int64) (appmodel.Configuration, error) {
	dbConfig, err := dbgen.ConfigurationQuery().
		Where(dbgen.ConfigurationWhere.ID.EQ(configID)).
		FetchOne(ctx, db)

	if errors.Is(err, sql.ErrNoRows) {
		return appmodel.Configuration{}, ErrNotFound
	}
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("fetching config from database: %v", err)
	}

	appConfig, err := toAppConfig(dbConfig)
	if err != nil {
		return appmodel.Configuration{}, fmt.Errorf("creating App config from DB config: %v", err)
	}
	return appConfig, nil
}

func DeleteConfig(ctx context.Context, db bob.Executor, configID int64) error {
	_, err := dbgen.AssetQuery().
		Where(dbgen.AssetWhere.ConfigurationID.EQ(configID)).
		Delete(ctx, db)
	if err != nil {
		return fmt.Errorf("deleting assets from database: %v", err)
	}

	count, err := dbgen.ConfigurationQuery().
		Where(dbgen.ConfigurationWhere.ID.EQ(configID)).
		Delete(ctx, db)

	if err != nil {
		return fmt.Errorf("deleting config from database: %v", err)
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID", count)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func toDbConfig(ctx context.Context, appConfig appmodel.Configuration) (dbgen.ConfigurationSetter, error) {
	dbConfig := dbgen.ConfigurationSetter{
		ID:                appConfig.Id,
		APIAccessChangeMe: appConfig.ApiAccessChangeMe,
		Enable:            appConfig.Enable,
		RefreshInterval:   appConfig.RefreshInterval,
		RequestTimeout:    appConfig.RequestTimeout,
		Active:            appConfig.Active,
		ProjectIds:        appConfig.ProjectIDs,
	}

	env := frontend.GetEnvironment(ctx)
	if env != nil {
		dbConfig.UserID = env.UserId
	}

	return dbConfig, nil
}

func toAppConfig(dbConfig dbgen.Configuration) (appmodel.Configuration, error) {
	var af [][]appmodel.FilterRule
	if err := json.Unmarshal(dbConfig.AssetFilter, &af); err != nil {
		return appmodel.Configuration{}, fmt.Errorf("unmarshalling assetFilter: %v", err)
	}

	return appmodel.Configuration{
		ApiAccessChangeMe: dbConfig.APIAccessChangeMe,
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

func GetConfigs(ctx context.Context, db bob.Executor) ([]appmodel.Configuration, error) {
	dbConfigs, err := dbgen.ConfigurationQuery().Fetch(ctx, db)
	if err != nil {
		return nil, err
	}

	var appConfigs []appmodel.Configuration
	for _, dbConfig := range dbConfigs {
		ac, err := toAppConfig(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("creating App config from DB config: %v", err)
		}
		appConfigs = append(appConfigs, ac)
	}
	return appConfigs, nil
}

func SetConfigActiveState(ctx context.Context, db bob.Executor, config appmodel.Configuration, state bool) (int64, error) {
	return dbgen.ConfigurationQuery().
		Where(dbgen.ConfigurationWhere.ID.EQ(config.Id)).
		Update().
		Set(dbgen.ConfigurationColumns.Active, state).
		Exec(ctx, db)
}

func SetAllConfigsInactive(ctx context.Context, db bob.Executor) (int64, error) {
	return dbgen.ConfigurationQuery().
		Update().
		Set(dbgen.ConfigurationColumns.Active, false).
		Exec(ctx, db)
}

// Insert an asset into the database
func InsertAsset(ctx context.Context, db bob.Executor, config appmodel.Configuration, projId string, globalAssetID string, assetId int32, providerId string) error {
	dbAsset := dbgen.Asset{
		ConfigurationID: config.Id,
		ProjectID:       projId,
		GlobalAssetID:   globalAssetID,
		AssetID:         null.Int32From(assetId),
		ProviderID:      providerId,
	}

	return dbgen.InsertAsset(dbAsset).Exec(ctx, db)
}

// Retrieve an asset ID given configuration, project, and global asset ID
func GetAssetId(ctx context.Context, db bob.Executor, config appmodel.Configuration, projId string, globalAssetID string) (*int32, error) {
	dbAsset, err := dbgen.AssetQuery().
		Where(
			dbgen.AssetWhere.ConfigurationID.EQ(config.Id),
			dbgen.AssetWhere.ProjectID.EQ(projId),
			dbgen.AssetWhere.GlobalAssetID.EQ(globalAssetID),
		).
		FetchOne(ctx, db)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil pointer instead of error if no record found
		}
		return nil, err
	}

	return common.Ptr(dbAsset.AssetID.Int32), nil
}

// Convert a database asset into an application model asset
func toAppAsset(dbAsset dbgen.Asset, config appmodel.Configuration) appmodel.Asset {
	return appmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		ProjectID:     dbAsset.ProjectID,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
	}
}

// Fetch an asset by its ID, including its related configuration
func GetAssetById(ctx context.Context, db bob.Executor, assetId int32) (appmodel.Asset, error) {
	// Fetch asset with its related configuration
	asset, err := dbgen.AssetQuery().
		Where(dbgen.AssetWhere.ID.EQ(assetId)).
		Join(dbgen.AssetJoin.Configuration). // Equivalent to eager loading Configuration
		FetchOne(ctx, db)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appmodel.Asset{}, ErrNotFound
		}
		return appmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}

	if !asset.AssetID.Valid {
		return appmodel.Asset{}, fmt.Errorf("shouldn't happen: assetID is nil")
	}

	// Extract configuration
	if asset.Configuration == nil {
		return appmodel.Asset{}, fmt.Errorf("shouldn't happen: configuration is nil")
	}

	config, err := toAppConfig(*asset.Configuration)
	if err != nil {
		return appmodel.Asset{}, fmt.Errorf("translating configuration: %v", err)
	}

	return toAppAsset(asset, config), nil
}
