// Copyright 2018 The OpenSDS Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/opensds/opensds/pkg/context"
	"github.com/opensds/opensds/pkg/db"
	"github.com/opensds/opensds/pkg/model"
	. "github.com/opensds/opensds/testutils/collection"
	dbtest "github.com/opensds/opensds/testutils/db/testing"
)

var assertTestResult = func(t *testing.T, got, expected interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v\n", expected, got)
	}
}

func TestCreateVolumeDBEntry(t *testing.T) {
	var in = &model.VolumeSpec{
		BaseModel:   &model.BaseModel{},
		Name:        "volume sample",
		Description: "This is a sample volume for testing",
		Size:        int64(1),
		ProfileId:   "3769855c-a102-11e7-b772-17b880d2f537",
		Status:      model.VolumeCreating,
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[0], nil)
		db.C = mockClient

		var expected = &SampleVolumes[0]
		result, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		if err != nil {
			t.Errorf("failed to create volume asynchronously, err is %v\n", err)
		}
		assertTestResult(t, result, expected)
	})

	t.Run("The size of volume created should be greater than zero", func(t *testing.T) {
		in.Size = int64(-2)
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[0], nil)
		db.C = mockClient

		_, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		expectedError := fmt.Sprintf("invalid volume size: %d", in.Size)
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("The profile id should not be empty", func(t *testing.T) {
		in.Size, in.ProfileId = int64(1), ""
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[0], nil)
		db.C = mockClient

		_, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		expectedError := "profile id can not be empty when creating volume in db"
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestCreateVolumeFromSnapshotDBEntry(t *testing.T) {
	var in = &model.VolumeSpec{
		BaseModel:   &model.BaseModel{},
		Name:        "volume sample",
		Description: "This is a sample volume for testing",
		Size:        int64(1),
		ProfileId:   "3769855c-a102-11e7-b772-17b880d2f537",
		Status:      model.VolumeCreating,
		SnapshotId:  "3769855c-a102-11e7-b772-17b880d2f537",
	}
	var snap = &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		Size:   int64(1),
		Status: model.VolumeSnapAvailable,
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
		mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
		db.C = mockClient

		var expected = &SampleVolumes[1]
		result, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		if err != nil {
			t.Errorf("failed to create volume with snapshot, err is %v\n", err)
		}
		assertTestResult(t, result, expected)
	})

	t.Run("The status of volume snapshot should always be available", func(t *testing.T) {
		snap.Status = model.VolumeSnapError
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
		mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
		db.C = mockClient

		_, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		expectedError := "only if the snapshot is available, the volume can be created"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("Size of volume should always be equal to or bigger than size of the snapshot", func(t *testing.T) {
		snap.Status, snap.Size = model.VolumeSnapAvailable, 10
		mockClient := new(dbtest.Client)
		mockClient.On("CreateVolume", context.NewAdminContext(), in).Return(&SampleVolumes[1], nil)
		mockClient.On("GetVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537").Return(snap, nil)
		db.C = mockClient

		_, err := CreateVolumeDBEntry(context.NewAdminContext(), in)
		expectedError := "size of volume must be equal to or bigger than size of the snapshot"
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestDeleteVolumeDBEntry(t *testing.T) {
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.VolumeAvailable,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}
	var in = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.VolumeDeleting,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
		mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
		mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
		mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		err := DeleteVolumeDBEntry(context.NewAdminContext(), vol)
		if err != nil {
			t.Errorf("failed to delete volume, err is %v\n", err)
		}
	})

	t.Run("Volume to be deleted should not contain any snapshots", func(t *testing.T) {
		var sampleSnapshots = []*model.VolumeSnapshotSpec{&SampleSnapshots[0]}
		// Considering vol has been updated inisde DeleteVolumeDBEntry, so the status
		// should be rolled back here.
		vol.Status = model.VolumeAvailable
		mockClient := new(dbtest.Client)
		mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
		mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(sampleSnapshots, nil)
		mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
		mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		err := DeleteVolumeDBEntry(context.NewAdminContext(), vol)
		expectedError := fmt.Sprintf("volume %s can not be deleted, because it still has snapshots", in.Id)
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("Volume to be deleted should not be in-use", func(t *testing.T) {
		var sampleAttachments = []*model.VolumeAttachmentSpec{&SampleAttachments[0]}
		// Considering vol has been updated inisde DeleteVolumeDBEntry, so the status
		// should be rolled back here.
		vol.Status = model.VolumeAvailable
		mockClient := new(dbtest.Client)
		mockClient.On("DeleteVolume", context.NewAdminContext(), vol.Id).Return(nil)
		mockClient.On("ListSnapshotsByVolumeId", context.NewAdminContext(), vol.Id).Return(nil, nil)
		mockClient.On("ListAttachmentsByVolumeId", context.NewAdminContext(), vol.Id).Return(sampleAttachments, nil)
		mockClient.On("UpdateVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		err := DeleteVolumeDBEntry(context.NewAdminContext(), vol)
		expectedError := fmt.Sprintf("volume %s can not be deleted, because it's in use", in.Id)
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestExtendVolumeDBEntry(t *testing.T) {
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: model.VolumeAvailable,
		Size:   2,
	}
	var in = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: model.VolumeExtending,
		Size:   2,
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
		mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		_, err := ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 20})
		if err != nil {
			t.Errorf("failed to extend volume: %v\n", err)
		}
	})

	t.Run("The status of volume should always be available", func(t *testing.T) {
		vol.Status = model.VolumeCreating
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
		mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		_, err := ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 20})
		expectedError := "the status of the volume to be extended must be available!"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("The extended size should always be larger than current size", func(t *testing.T) {
		vol.Size, vol.Status = 20, model.VolumeAvailable
		in.Size = 20
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
		mockClient.On("ExtendVolume", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		_, err := ExtendVolumeDBEntry(context.NewAdminContext(), vol.Id, &model.ExtendVolumeSpec{NewSize: 2})
		expectedError := "new size for extend must be greater than current size." +
			"(current: 20 GB, extended: 2 GB)."
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestCreateVolumeAttachmentDBEntry(t *testing.T) {
	var req = &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{},
		VolumeId:  "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Status:    "creating",
	}

	t.Run("Volume status should be available that attachment can be created", func(t *testing.T) {
		var vol1 = &model.VolumeSpec{
			BaseModel: &model.BaseModel{
				Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			},
			Status: "error",
		}

		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol1, nil)
		db.C = mockClient

		_, err := CreateVolumeAttachmentDBEntry(context.NewAdminContext(), req)
		expectedError := "only the status of volume is available, attachment can be created"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("If volume status is in-use, the multi-attach should be true, attachment can be created", func(t *testing.T) {
		var vol2 = &model.VolumeSpec{
			BaseModel: &model.BaseModel{
				Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			},
			Status:      "inUse",
			MultiAttach: false,
		}
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol2, nil)
		db.C = mockClient

		_, err := CreateVolumeAttachmentDBEntry(context.NewAdminContext(), req)
		expectedError := "volume is already attached or volume multiattach must be true if attach more than once"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("Volume status is in-use and multi-attach is true, attachment created successfully", func(t *testing.T) {
		var vol3 = &model.VolumeSpec{
			BaseModel: &model.BaseModel{
				Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			},
			Status:      "inUse",
			MultiAttach: true,
		}
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol3, nil)
		mockClient.On("UpdateStatus", context.NewAdminContext(), vol3, "attaching").Return(nil)
		mockClient.On("CreateVolumeAttachment", context.NewAdminContext(), req).Return(&SampleAttachments[0], nil)
		db.C = mockClient

		var expected = &SampleAttachments[0]

		result, _ := CreateVolumeAttachmentDBEntry(context.NewAdminContext(), req)
		assertTestResult(t, result, expected)
	})

	t.Run("Volume status is available, attachment created successfully", func(t *testing.T) {
		var vol4 = &model.VolumeSpec{
			BaseModel: &model.BaseModel{
				Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
			},
			Status: "available",
		}
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol4, nil)
		mockClient.On("UpdateStatus", context.NewAdminContext(), vol4, "attaching").Return(nil)
		mockClient.On("CreateVolumeAttachment", context.NewAdminContext(), req).Return(&SampleAttachments[0], nil)
		db.C = mockClient

		expected := &SampleAttachments[0]
		result, _ := CreateVolumeAttachmentDBEntry(context.NewAdminContext(), req)
		assertTestResult(t, result, expected)
	})
}

func TestDeleteVolumeAttachmentDBEntry(t *testing.T) {
	var req = &model.VolumeAttachmentSpec{
		BaseModel: &model.BaseModel{
			Id: "f2dda3d2-bf79-11e7-8665-f750b088f63e",
		},
		VolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Status:   "available",
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("UpdateVolumeAttachment", context.NewAdminContext(), "f2dda3d2-bf79-11e7-8665-f750b088f63e", req).Return(nil, nil)
		mockClient.On("GetVolume", context.NewAdminContext(), req.VolumeId).Return(nil, nil)
		db.C = mockClient

		err := DeleteVolumeAttachmentDBEntry(context.NewAdminContext(), req)
		if err != nil {
			t.Errorf("failed to delete volume attachment, err is %v\n", err)
		}
	})
}

func TestCreateVolumeSnapshotDBEntry(t *testing.T) {
	var vol = &model.VolumeSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Size:   1,
		Status: "available",
	}
	var req = &model.VolumeSnapshotSpec{
		BaseModel:   &model.BaseModel{},
		VolumeId:    "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Name:        "sample-snapshot-01",
		Description: "This is the first sample snapshot for testing",
		Size:        int64(1),
		ProfileId:   "3769855c-a102-11e7-b772-17b880d2f537",
		Status:      "creating",
		Metadata:    map[string]string{"a": "a"},
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
		mockClient.On("CreateVolumeSnapshot", context.NewAdminContext(), req).Return(&SampleSnapshots[0], nil)
		db.C = mockClient

		var expected = &SampleSnapshots[0]
		result, err := CreateVolumeSnapshotDBEntry(context.NewAdminContext(), req)
		if err != nil {
			t.Errorf("failed to create volume snapshot, err is %v\n", err)
		}
		assertTestResult(t, result, expected)
	})

	t.Run("The profile id should not be empty", func(t *testing.T) {
		req.ProfileId = ""
		mockClient := new(dbtest.Client)
		mockClient.On("GetVolume", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(vol, nil)
		mockClient.On("CreateVolumeSnapshot", context.NewAdminContext(), req).Return(&SampleSnapshots[0], nil)
		db.C = mockClient

		_, err := CreateVolumeSnapshotDBEntry(context.NewAdminContext(), req)
		expectedError := "profile id can not be empty when creating volume snapshot in db"
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestDeleteVolumeSnapshotDBEntry(t *testing.T) {
	var req = &model.VolumeSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		VolumeId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Status:   "available",
	}

	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("UpdateVolumeSnapshot", context.NewAdminContext(), "3769855c-a102-11e7-b772-17b880d2f537", req).Return(nil, nil)
		mockClient.On("GetVolume", context.NewAdminContext(), req.VolumeId).Return(nil, nil)
		db.C = mockClient

		err := DeleteVolumeSnapshotDBEntry(context.NewAdminContext(), req)
		if err != nil {
			t.Errorf("failed to delete volume snapshot, err is %v\n", err)
		}
	})
}

func TestCreateFileShareSnapshotDBEntry(t *testing.T) {
	var fileshare = &model.FileShareSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status: "available",
	}
	var req = &model.FileShareSnapshotSpec{
		BaseModel: &model.BaseModel{
			Id: "3769855c-a102-11e7-b772-17b880d2f537",
		},
		Name:        "sample-snapshot-01",
		Description: "This is the first sample snapshot for testing",
		Status:      "available",
		ShareSize:  int64(1),
		FileShareId: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		ProfileId:   "1106b972-66ef-11e7-b172-db03f3689c9c",
	}

	var sampleSnapshots = []*model.FileShareSnapshotSpec{&SampleShareSnapshots[0]}
	t.Run("-ve test case - snapshot name already exists", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("GetFileShare", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(fileshare, nil)
		mockClient.On("ListFileShareSnapshots", context.NewAdminContext()).Return(sampleSnapshots, nil)
		db.C = mockClient

		_, err := CreateFileShareSnapshotDBEntry(context.NewAdminContext(), req)
		expectedError := "file share snapshot name already exists"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("test +ve", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("GetFileShare", context.NewAdminContext(), "bd5b12a8-a101-11e7-941e-d77981b584d8").Return(fileshare, nil)
		mockClient.On("ListFileShareSnapshots", context.NewAdminContext()).Return(nil, nil)
		mockClient.On("CreateFileShareSnapshot", context.NewAdminContext(), req).Return(&SampleShareSnapshots[0], nil)
		db.C = mockClient

		var expected = &SampleShareSnapshots[0]
		result, err := CreateFileShareSnapshotDBEntry(context.NewAdminContext(), req)
		if err != nil {
			t.Errorf("failed to create fileshare snapshot, err is %v\n", err)
		}
		assertTestResult(t, result, expected)
	})

}

func TestCreateFileShareDBEntry(t *testing.T) {
	var in = &model.FileShareSpec{
		BaseModel:   &model.BaseModel{},
		Name:        "sample-fileshare-01",
		Description: "This is a sample fileshare for testing",
		Size:        int64(1),
		ProfileId:   "b3585ebe-c42c-120g-b28e-f373746a71ca",
		Status:      model.FileShareCreating,
	}
	
	t.Run("Everything should work well", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		var expected = &SampleFileShares[0]
		result, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		if err != nil {
			t.Errorf("failed to create fileshare err is %v\n", err)
		}
		assertTestResult(t, result, expected)
	})

	t.Run("The size of fileshare created should be greater than zero", func(t *testing.T) {
		in.Size = int64(-2)
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		_, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		expectedError := fmt.Sprintf("invalid fileshare size: %d", in.Size)
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("The profile id should not be empty", func(t *testing.T) {
		in.ProfileId = ""
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		_, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		expectedError := "profile id can not be empty when creating fileshare in db!"
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("Empty file share name is allowed", func(t *testing.T) {
		in.Size, in.Name, in.ProfileId = int64(1),"", "b3585ebe-c42c-120g-b28e-f373746a71ca"
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		_, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		expectedError := "empty fileshare name is not allowed. Please give valid name."
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("File share name length more than 128 characters are not allowed", func(t *testing.T) {
		in.Size, in.Name, in.ProfileId = int64(1),"kjxdpvfuenecpyhbxdumwibipjkigcjykabuhghjghjgjgjgjgjghhkkkkkkkkkkkkkkklsccdesnkgxnjnhfjusigceorogdgjlyciemscgvncerucokfekdiviuyplbly", "b3585ebe-c42c-120g-b28e-f373746a71ca"
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		_, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		expectedError := "fileshare name length should not more than 128 characters. current length is : "+strconv.Itoa(len(in.Name))
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("Special characters in file share description are not allowed", func(t *testing.T) {
		in.Size, in.Name, in.ProfileId, in.Description = int64(1),"sample-fileshare-01", "b3585ebe-c42c-120g-b28e-f373746a71ca", "#FileShare Code!$! test"
		mockClient := new(dbtest.Client)
		mockClient.On("CreateFileShare", context.NewAdminContext(), in).Return(&SampleFileShares[0], nil)
		db.C = mockClient

		_, err := CreateFileShareDBEntry(context.NewAdminContext(), in)
		expectedError := "invalid fileshare description and it has some special characters"
		assertTestResult(t, err.Error(), expectedError)
	})
}

func TestDeleteFileShareDBEntry(t *testing.T) {
	var fileshare = &model.FileShareSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.FileShareAvailable,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}
	var in = &model.FileShareSpec{
		BaseModel: &model.BaseModel{
			Id: "bd5b12a8-a101-11e7-941e-d77981b584d8",
		},
		Status:    model.FileShareInUse,
		ProfileId: "3769855c-a102-11e7-b772-17b880d2f537",
		PoolId:    "3762355c-a102-11e7-b772-17b880d2f537",
	}
	t.Run("FileShare to be deleted should not be in-use", func(t *testing.T) {
		fileshare.Status = model.FileShareInUse
		mockClient := new(dbtest.Client)
		mockClient.On("ListSnapshotsByShareId", context.NewAdminContext(), fileshare.Id).Return(nil, nil)
		mockClient.On("ListFileShareAclsByShareId", context.NewAdminContext(), fileshare.Id).Return(nil, nil)
		mockClient.On("UpdateFileShare", context.NewAdminContext(), in).Return(nil, nil)
		mockClient.On("DeleteFileShare", context.NewAdminContext(), fileshare.Id).Return(nil)
		db.C = mockClient

		err := DeleteFileShareDBEntry(context.NewAdminContext(), fileshare)
		expectedError := fmt.Sprintf("only the fileshare with the status available, error, errorDeleting, can be deleted, the fileshare status is %s", in.Status)
		assertTestResult(t, err.Error(), expectedError)
	})

	var sampleSnapshots = []*model.FileShareSnapshotSpec{&SampleShareSnapshots[0]}
	t.Run("FileShare should not be deleted if it has dependent snapshots", func(t *testing.T) {
		//in.Status = model.FileShareAvailable
		fileshare.Status = model.FileShareAvailable
		mockClient := new(dbtest.Client)
		mockClient.On("ListSnapshotsByShareId", context.NewAdminContext(), fileshare.Id).Return(sampleSnapshots, nil)
		db.C = mockClient

		err := DeleteFileShareDBEntry(context.NewAdminContext(), fileshare)
		expectedError := fmt.Sprintf("file share %s can not be deleted, because it still has snapshots", in.Id)
		assertTestResult(t, err.Error(), expectedError)
	})

	var sampleAcls = []*model.FileShareAclSpec{&SampleFileSharesAcl[2]}
	t.Run("FileShare should not be deleted if it has dependent acls", func(t *testing.T) {
		//in.Status = model.FileShareAvailable
		fileshare.Status = model.FileShareAvailable
		mockClient := new(dbtest.Client)
		mockClient.On("ListSnapshotsByShareId", context.NewAdminContext(), fileshare.Id).Return(nil, nil)
		mockClient.On("ListFileShareAclsByShareId", context.NewAdminContext(), fileshare.Id).Return(sampleAcls, nil)
		db.C = mockClient

		err := DeleteFileShareDBEntry(context.NewAdminContext(), fileshare)
		expectedError := fmt.Sprintf("file share %s can not be deleted, because it still has acls", in.Id)
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("FileShare deletion when it is available", func(t *testing.T) {
		in.Status = model.FileShareDeleting
		//fileshare.Status = model.FileShareAvailable
		mockClient := new(dbtest.Client)
		mockClient.On("ListSnapshotsByShareId", context.NewAdminContext(), fileshare.Id).Return(nil, nil)
		mockClient.On("ListFileShareAclsByShareId", context.NewAdminContext(), fileshare.Id).Return(nil, nil)
		mockClient.On("UpdateFileShare", context.NewAdminContext(), in).Return(nil, nil)
		mockClient.On("DeleteFileShare", context.NewAdminContext(), fileshare.Id).Return(nil)
		db.C = mockClient

		err := DeleteFileShareDBEntry(context.NewAdminContext(), fileshare)
		if err != nil {
			t.Errorf("failed to delete fileshare, err is %v\n", err)
		}
	})
}

func TestDeleteFileShareAclDBEntry(t *testing.T) {
	var in = &model.FileShareAclSpec{
		BaseModel: &model.BaseModel{
			Id: "d2975ebe-d82c-430f-b28e-f373746a71ca",
		},
		Status:           model.FileShareAclAvailable,
		ProfileId:        "3769855c-a102-11e7-b772-17b880d2f537",
		FileShareId:      "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Type:             "ip",
		AccessTo:         "10.21.23.10",
		AccessCapability: []string{"Read", "Write"},
	}
	var out = &model.FileShareAclSpec{
		BaseModel: &model.BaseModel{
			Id: "d2975ebe-d82c-430f-b28e-f373746a71ca",
		},
		Status:           model.FileShareAclDeleting,
		ProfileId:        "3769855c-a102-11e7-b772-17b880d2f537",
		FileShareId:      "bd5b12a8-a101-11e7-941e-d77981b584d8",
		Type:             "ip",
		AccessTo:         "10.21.23.10",
		AccessCapability: []string{"Read", "Write"},
	}

	t.Run("FileShareAcl to be deleted should not be in-use", func(t *testing.T) {
		in.Status = model.FileShareAclInUse
		mockClient := new(dbtest.Client)
		mockClient.On("GetFileShare", context.NewAdminContext(), in.FileShareId).Return(nil, nil)
		mockClient.On("DeleteFileShareAcl", context.NewAdminContext(), in.Id).Return(nil, nil)
		mockClient.On("UpdateFileShareAcl", context.NewAdminContext(), in).Return(nil, nil)
		db.C = mockClient

		err := DeleteFileShareAclDBEntry(context.NewAdminContext(), in)
		expectedError := fmt.Sprintf("only the file share acl with the status available, error, error_deleting can be deleted, the fileshare status is %s", in.Status)
		assertTestResult(t, err.Error(), expectedError)
	})

	t.Run("FileShareAcl deletion when everything works fine", func(t *testing.T) {
		mockClient := new(dbtest.Client)
		in.Status = model.FileShareAclAvailable
		mockClient.On("GetFileShare", context.NewAdminContext(), in.FileShareId).Return(&SampleFileShares[0], nil)
		mockClient.On("DeleteFileShareAcl", context.NewAdminContext(), in.Id).Return(nil, nil)
		mockClient.On("UpdateFileShareAcl", context.NewAdminContext(), in).Return(out, nil)
		db.C = mockClient

		err := DeleteFileShareAclDBEntry(context.NewAdminContext(), in)
		if err != nil {
			t.Errorf("failed delete fileshare acl in db:%v\n", err)
		}
	})
}
