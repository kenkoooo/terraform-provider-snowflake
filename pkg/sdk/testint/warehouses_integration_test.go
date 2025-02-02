package testint

import (
	"testing"
	"time"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt_WarehousesShow(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)

	// new warehouses created on purpose
	testWarehouse, warehouseCleanup := createWarehouseWithOptions(t, client, &sdk.CreateWarehouseOptions{
		WarehouseSize: &sdk.WarehouseSizeSmall,
	})
	t.Cleanup(warehouseCleanup)
	_, warehouse2Cleanup := createWarehouse(t, client)
	t.Cleanup(warehouse2Cleanup)

	t.Run("show without options", func(t *testing.T) {
		warehouses, err := client.Warehouses.Show(ctx, nil)
		require.NoError(t, err)
		assert.LessOrEqual(t, 2, len(warehouses))
	})

	t.Run("show with options", func(t *testing.T) {
		showOptions := &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: &testWarehouse.Name,
			},
		}
		warehouses, err := client.Warehouses.Show(ctx, showOptions)
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		assert.Equal(t, testWarehouse.Name, warehouses[0].Name)
		assert.Equal(t, sdk.WarehouseSizeSmall, warehouses[0].Size)
	})

	t.Run("when searching a non-existent password policy", func(t *testing.T) {
		showOptions := &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String("non-existent"),
			},
		}
		warehouses, err := client.Warehouses.Show(ctx, showOptions)
		require.NoError(t, err)
		assert.Equal(t, 0, len(warehouses))
	})
}

func TestInt_WarehouseCreate(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)
	tagTest, tagCleanup := createTag(t, client, testDb(t), testSchema(t))
	t.Cleanup(tagCleanup)
	tag2Test, tag2Cleanup := createTag(t, client, testDb(t), testSchema(t))
	t.Cleanup(tag2Cleanup)

	t.Run("test complete", func(t *testing.T) {
		id := sdk.RandomAccountObjectIdentifier()
		err := client.Warehouses.Create(ctx, id, &sdk.CreateWarehouseOptions{
			OrReplace:                       sdk.Bool(true),
			WarehouseType:                   &sdk.WarehouseTypeStandard,
			WarehouseSize:                   &sdk.WarehouseSizeSmall,
			MaxClusterCount:                 sdk.Int(8),
			MinClusterCount:                 sdk.Int(2),
			ScalingPolicy:                   &sdk.ScalingPolicyEconomy,
			AutoSuspend:                     sdk.Int(1000),
			AutoResume:                      sdk.Bool(true),
			InitiallySuspended:              sdk.Bool(false),
			Comment:                         sdk.String("comment"),
			EnableQueryAcceleration:         sdk.Bool(true),
			QueryAccelerationMaxScaleFactor: sdk.Int(90),
			MaxConcurrencyLevel:             sdk.Int(10),
			StatementQueuedTimeoutInSeconds: sdk.Int(2000),
			StatementTimeoutInSeconds:       sdk.Int(3000),
			Tag: []sdk.TagAssociation{
				{
					Name:  tagTest.ID(),
					Value: "v1",
				},
				{
					Name:  tag2Test.ID(),
					Value: "v2",
				},
			},
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			err = client.Warehouses.Drop(ctx, id, &sdk.DropWarehouseOptions{
				IfExists: sdk.Bool(true),
			})
			require.NoError(t, err)
		})
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(id.Name()),
			},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(warehouses))
		warehouse := warehouses[0]
		assert.Equal(t, id.Name(), warehouse.Name)
		assert.Equal(t, sdk.WarehouseTypeStandard, warehouse.Type)
		assert.Equal(t, sdk.WarehouseSizeSmall, warehouse.Size)
		assert.Equal(t, 8, warehouse.MaxClusterCount)
		assert.Equal(t, 2, warehouse.MinClusterCount)
		assert.Equal(t, sdk.ScalingPolicyEconomy, warehouse.ScalingPolicy)
		assert.Equal(t, 1000, warehouse.AutoSuspend)
		assert.Equal(t, true, warehouse.AutoResume)
		assert.Contains(t, []sdk.WarehouseState{sdk.WarehouseStateResuming, sdk.WarehouseStateStarted}, warehouse.State)
		assert.Equal(t, "comment", warehouse.Comment)
		assert.Equal(t, true, warehouse.EnableQueryAcceleration)
		assert.Equal(t, 90, warehouse.QueryAccelerationMaxScaleFactor)

		tag1Value, err := client.SystemFunctions.GetTag(ctx, tagTest.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		assert.Equal(t, "v1", tag1Value)
		tag2Value, err := client.SystemFunctions.GetTag(ctx, tag2Test.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		assert.Equal(t, "v2", tag2Value)
	})

	t.Run("test no options", func(t *testing.T) {
		id := sdk.RandomAccountObjectIdentifier()
		err := client.Warehouses.Create(ctx, id, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = client.Warehouses.Drop(ctx, id, &sdk.DropWarehouseOptions{
				IfExists: sdk.Bool(true),
			})
			require.NoError(t, err)
		})
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(id.Name()),
			},
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		assert.Equal(t, id.Name(), result.Name)
		assert.Equal(t, sdk.WarehouseTypeStandard, result.Type)
		assert.Equal(t, sdk.WarehouseSizeXSmall, result.Size)
		assert.Equal(t, 1, result.MaxClusterCount)
		assert.Equal(t, 1, result.MinClusterCount)
		assert.Equal(t, sdk.ScalingPolicyStandard, result.ScalingPolicy)
		assert.Equal(t, 600, result.AutoSuspend)
		assert.Equal(t, true, result.AutoResume)
		assert.Contains(t, []sdk.WarehouseState{sdk.WarehouseStateResuming, sdk.WarehouseStateStarted}, result.State)
		assert.Equal(t, "", result.Comment)
		assert.Equal(t, false, result.EnableQueryAcceleration)
		assert.Equal(t, 8, result.QueryAccelerationMaxScaleFactor)
	})
}

func TestInt_WarehouseDescribe(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)

	// new warehouse created on purpose
	warehouse, warehouseCleanup := createWarehouse(t, client)
	t.Cleanup(warehouseCleanup)

	t.Run("when warehouse exists", func(t *testing.T) {
		result, err := client.Warehouses.Describe(ctx, warehouse.ID())
		require.NoError(t, err)
		assert.Equal(t, warehouse.Name, result.Name)
		assert.Equal(t, "WAREHOUSE", result.Kind)
		assert.WithinDuration(t, time.Now(), result.CreatedOn, 5*time.Second)
	})

	t.Run("when warehouse does not exist", func(t *testing.T) {
		id := sdk.NewAccountObjectIdentifier("does_not_exist")
		_, err := client.Warehouses.Describe(ctx, id)
		assert.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)
	})
}

func TestInt_WarehouseAlter(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)

	tag, tagCleanup := createTag(t, client, testDb(t), testSchema(t))
	t.Cleanup(tagCleanup)
	tag2, tagCleanup2 := createTag(t, client, testDb(t), testSchema(t))
	t.Cleanup(tagCleanup2)

	t.Run("terraform acc test", func(t *testing.T) {
		id := sdk.RandomAccountObjectIdentifier()
		opts := &sdk.CreateWarehouseOptions{
			Comment:            sdk.String("test comment"),
			WarehouseSize:      &sdk.WarehouseSizeXSmall,
			AutoSuspend:        sdk.Int(60),
			MaxClusterCount:    sdk.Int(1),
			MinClusterCount:    sdk.Int(1),
			ScalingPolicy:      &sdk.ScalingPolicyStandard,
			AutoResume:         sdk.Bool(true),
			InitiallySuspended: sdk.Bool(true),
		}
		err := client.Warehouses.Create(ctx, id, opts)
		require.NoError(t, err)
		t.Cleanup(func() {
			err = client.Warehouses.Drop(ctx, id, &sdk.DropWarehouseOptions{
				IfExists: sdk.Bool(true),
			})
			require.NoError(t, err)
		})
		warehouse, err := client.Warehouses.ShowByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, 1, warehouse.MaxClusterCount)
		assert.Equal(t, 1, warehouse.MinClusterCount)
		assert.Equal(t, sdk.ScalingPolicyStandard, warehouse.ScalingPolicy)
		assert.Equal(t, 60, warehouse.AutoSuspend)
		assert.Equal(t, true, warehouse.AutoResume)
		assert.Equal(t, "test comment", warehouse.Comment)
		assert.Equal(t, sdk.WarehouseStateSuspended, warehouse.State)
		assert.Equal(t, sdk.WarehouseSizeXSmall, warehouse.Size)

		// rename
		newID := sdk.RandomAccountObjectIdentifier()
		alterOptions := &sdk.AlterWarehouseOptions{
			NewName: &newID,
		}
		err = client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouse, err = client.Warehouses.ShowByID(ctx, newID)
		require.NoError(t, err)
		assert.Equal(t, newID.Name(), warehouse.Name)

		// change props
		alterOptions = &sdk.AlterWarehouseOptions{
			Set: &sdk.WarehouseSet{
				WarehouseSize: &sdk.WarehouseSizeSmall,
				Comment:       sdk.String("test comment2"),
			},
		}
		err = client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouse, err = client.Warehouses.ShowByID(ctx, newID)
		require.NoError(t, err)
		assert.Equal(t, "test comment2", warehouse.Comment)
		assert.Equal(t, sdk.WarehouseSizeSmall, warehouse.Size)
	})

	t.Run("set", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		alterOptions := &sdk.AlterWarehouseOptions{
			Set: &sdk.WarehouseSet{
				WarehouseSize:           &sdk.WarehouseSizeMedium,
				AutoSuspend:             sdk.Int(1234),
				EnableQueryAcceleration: sdk.Bool(true),
			},
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		assert.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		require.NoError(t, err)
		assert.Equal(t, sdk.WarehouseSizeMedium, result.Size)
		assert.Equal(t, true, result.EnableQueryAcceleration)
		assert.Equal(t, 1234, result.AutoSuspend)
	})

	t.Run("rename", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		oldID := warehouse.ID()
		t.Cleanup(warehouseCleanup)

		newID := sdk.RandomAccountObjectIdentifier()
		alterOptions := &sdk.AlterWarehouseOptions{
			NewName: &newID,
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		result, err := client.Warehouses.Describe(ctx, newID)
		require.NoError(t, err)
		assert.Equal(t, newID.Name(), result.Name)

		// rename back to original name so it can be cleaned up
		alterOptions = &sdk.AlterWarehouseOptions{
			NewName: &oldID,
		}
		err = client.Warehouses.Alter(ctx, newID, alterOptions)
		require.NoError(t, err)
	})

	t.Run("unset", func(t *testing.T) {
		createOptions := &sdk.CreateWarehouseOptions{
			Comment:         sdk.String("test comment"),
			MaxClusterCount: sdk.Int(10),
		}
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouseWithOptions(t, client, createOptions)
		t.Cleanup(warehouseCleanup)
		id := warehouse.ID()

		alterOptions := &sdk.AlterWarehouseOptions{
			Unset: &sdk.WarehouseUnset{
				Comment:         sdk.Bool(true),
				MaxClusterCount: sdk.Bool(true),
			},
		}
		err := client.Warehouses.Alter(ctx, id, alterOptions)
		require.NoError(t, err)
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		assert.Equal(t, warehouse.Name, result.Name)
		assert.Equal(t, "", result.Comment)
		assert.Equal(t, 1, result.MaxClusterCount)
	})

	t.Run("suspend & resume", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		alterOptions := &sdk.AlterWarehouseOptions{
			Suspend: sdk.Bool(true),
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		assert.Contains(t, []sdk.WarehouseState{sdk.WarehouseStateSuspended, sdk.WarehouseStateSuspending}, result.State)

		alterOptions = &sdk.AlterWarehouseOptions{
			Resume: sdk.Bool(true),
		}
		err = client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouses, err = client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result = warehouses[0]
		assert.Contains(t, []sdk.WarehouseState{sdk.WarehouseStateStarted, sdk.WarehouseStateResuming}, result.State)
	})

	t.Run("resume without suspending", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		alterOptions := &sdk.AlterWarehouseOptions{
			Resume:      sdk.Bool(true),
			IfSuspended: sdk.Bool(true),
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		assert.Contains(t, []sdk.WarehouseState{sdk.WarehouseStateStarted, sdk.WarehouseStateResuming}, result.State)
	})

	t.Run("abort all queries", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		resetWarehouse := useWarehouse(t, client, warehouse.ID())
		t.Cleanup(resetWarehouse)

		// Start a long query
		go client.ExecForTests(ctx, "CALL SYSTEM$WAIT(30);") //nolint:errcheck // we don't care if this eventually errors, as long as it runs for a little while
		time.Sleep(5 * time.Second)

		// Check that query is running
		warehouses, err := client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result := warehouses[0]
		assert.Equal(t, 1, result.Running)
		assert.Equal(t, 0, result.Queued)

		// Abort all queries
		alterOptions := &sdk.AlterWarehouseOptions{
			AbortAllQueries: sdk.Bool(true),
		}
		err = client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)

		// Wait for abort to be effective
		time.Sleep(5 * time.Second)

		// Check no query is running
		warehouses, err = client.Warehouses.Show(ctx, &sdk.ShowWarehouseOptions{
			Like: &sdk.Like{
				Pattern: sdk.String(warehouse.Name),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(warehouses))
		result = warehouses[0]
		assert.Equal(t, 0, result.Running)
		assert.Equal(t, 0, result.Queued)
	})

	t.Run("set tags", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		alterOptions := &sdk.AlterWarehouseOptions{
			Set: &sdk.WarehouseSet{
				Tag: []sdk.TagAssociation{
					{
						Name:  tag.ID(),
						Value: "val",
					},
					{
						Name:  tag2.ID(),
						Value: "val2",
					},
				},
			},
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)

		val, err := client.SystemFunctions.GetTag(ctx, tag.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		require.Equal(t, "val", val)
		val, err = client.SystemFunctions.GetTag(ctx, tag2.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		require.Equal(t, "val2", val)
	})

	t.Run("unset tags", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, warehouseCleanup := createWarehouse(t, client)
		t.Cleanup(warehouseCleanup)

		alterOptions := &sdk.AlterWarehouseOptions{
			Set: &sdk.WarehouseSet{
				Tag: []sdk.TagAssociation{
					{
						Name:  tag.ID(),
						Value: "val1",
					},
					{
						Name:  tag2.ID(),
						Value: "val2",
					},
				},
			},
		}
		err := client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)
		val, err := client.SystemFunctions.GetTag(ctx, tag.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		require.Equal(t, "val1", val)
		val2, err := client.SystemFunctions.GetTag(ctx, tag2.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.NoError(t, err)
		require.Equal(t, "val2", val2)

		alterOptions = &sdk.AlterWarehouseOptions{
			Unset: &sdk.WarehouseUnset{
				Tag: []sdk.ObjectIdentifier{
					tag.ID(),
					tag2.ID(),
				},
			},
		}
		err = client.Warehouses.Alter(ctx, warehouse.ID(), alterOptions)
		require.NoError(t, err)

		val, err = client.SystemFunctions.GetTag(ctx, tag.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.Error(t, err)
		require.Equal(t, "", val)
		val2, err = client.SystemFunctions.GetTag(ctx, tag2.ID(), warehouse.ID(), sdk.ObjectTypeWarehouse)
		require.Error(t, err)
		require.Equal(t, "", val2)
	})
}

func TestInt_WarehouseDrop(t *testing.T) {
	client := testClient(t)
	ctx := testContext(t)

	t.Run("when warehouse exists", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, _ := createWarehouse(t, client)

		err := client.Warehouses.Drop(ctx, warehouse.ID(), nil)
		require.NoError(t, err)
		_, err = client.Warehouses.Describe(ctx, warehouse.ID())
		assert.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)
	})

	t.Run("when warehouse does not exist", func(t *testing.T) {
		id := sdk.NewAccountObjectIdentifier("does_not_exist")
		err := client.Warehouses.Drop(ctx, id, nil)
		assert.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)
	})

	t.Run("when warehouse exists and if exists is true", func(t *testing.T) {
		// new warehouse created on purpose
		warehouse, _ := createWarehouse(t, client)

		dropOptions := &sdk.DropWarehouseOptions{IfExists: sdk.Bool(true)}
		err := client.Warehouses.Drop(ctx, warehouse.ID(), dropOptions)
		require.NoError(t, err)
		_, err = client.Warehouses.Describe(ctx, warehouse.ID())
		assert.ErrorIs(t, err, sdk.ErrObjectNotExistOrAuthorized)
	})
}
