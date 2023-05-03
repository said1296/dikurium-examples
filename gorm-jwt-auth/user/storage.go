package user

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"server/internal/constant"
	"server/internal/storage"
	"strconv"
	"strings"
)

func UpdateProfileImage(userID int, filename string, image io.Reader) error {
	profileImage, err := ProfileImagePath(userID)
	if err != nil {
		return err
	} else if profileImage != nil {
		err := os.Remove(*profileImage)
		if err != nil {
			return errors.Wrap(err, "failed to delete profile image")
		}
	}


	path := filepath.Join(constant.StorageUserPath, strconv.Itoa(userID), constant.StorageProfileImageName+filepath.Ext(filename))

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to create upload directories")
	}

	imageBuffer := new(bytes.Buffer)
	_, err = imageBuffer.ReadFrom(image)
	if err != nil {
		return errors.Wrap(err, "failed to convert image to bytes array")
	}

	err = storage.SaveFile(path, imageBuffer.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to save profile image")
	}

	return nil
}

func ProfileImagePath(userID int) (*string, error) {
	matches, err := filepath.Glob(filepath.Join(constant.StorageUserPath, strconv.Itoa(userID), constant.StorageProfileImageName) + ".*")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get profile image")
	} else if len(matches) == 0 {
		return nil, nil
	} else {
		return &matches[0], nil
	}
}

func ProfileImageURL(userID int) (*string, error) {
	profileImage, err := ProfileImagePath(userID)
	if err != nil {
		return nil, err
	} else if profileImage != nil {
		profileImageParsed := strings.Replace(*profileImage, constant.StorageUserPath, constant.AssetsUserURL, -1)
		profileImageParsed = filepath.ToSlash(filepath.Clean(profileImageParsed))
		profileImageParsed = strings.Replace(profileImageParsed, "..", "", -1)
		profileImage = &profileImageParsed
	}

	return profileImage, nil
}
