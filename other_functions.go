package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func saveImage(r *http.Request) (string, error) {
	// Parsing file
	file, stat, err := r.FormFile("image")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Checking if file size is lower than 20 Mb
	if stat.Size > 20*1024*1024 {
		return "", errors.New("Uploaded file size is too big! (Maximum 20 Mb)")
	}

	// Checking if file is corrupted or not and getting file stat
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", errors.New("Can't read your file, please try again")
	}

	//Checking if it is an image or not using file stat
	if !strings.HasPrefix(http.DetectContentType(buff), "image") {
		return "", errors.New("Only images can be uploaded")
	}

	// Checking cases when user sends image, but without extension.
	// This checker exists only because checking by stat says that it's an image
	var extension string
	extensionsArr := []string{".jpg", ".jpeg", ".jpe", ".jif", ".jfif", ".jfi", ".png", ".gif"}
	for _, ext := range extensionsArr {
		if strings.HasSuffix(stat.Filename, ext) {
			extension = ext
			break
		}
	}
	if extension == "" {
		return "", errors.New("Please add an extension to your image")
	}

	// Saving image
	fname, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/images/posts/%v%v", fname.String(), extension)
	newImage, err := os.Create("./db" + path)
	if err != nil {
		return "", err
	}
	file.Seek(0, 0)
	io.Copy(newImage, file)
	return path, nil
}

func getNewPostCategories(r *http.Request) ([]int, error) {
	cat1, err1 := strconv.Atoi(r.FormValue("categorie1"))
	cat2, _ := strconv.Atoi(r.FormValue("categorie2"))
	categories, err := getCategoriesList()
	if err != nil {
		return nil, err
	}
	lenCats := len(categories) + 1
	if err1 != nil ||
		cat1 < 0 || cat1 > lenCats ||
		cat2 < 0 || cat2 > lenCats {
		return nil, errors.New("Bad categorie ID")
	}
	return removeDuplicatesAndZeroes([]int{cat1, cat2}), nil
}

func removeDuplicatesAndZeroes(ints []int) []int {
	keys := make(map[int]bool)
	new := []int{}
	for _, n := range ints {
		if n == 0 {
			continue
		}
		if _, v := keys[n]; !v {
			keys[n] = true
			new = append(new, n)
		}
	}
	return new
}

func isEmpty(str string) bool {
	for _, v := range str {
		if v != ' ' && v != '	' && v != '\n' {
			return false
		}
	}
	return true
}

func isValidEmail(str string) bool {
	if isEmpty(str) {
		return false
	}

	if len(str) > 30 {
		return false
	}

	arr1 := strings.Split(str, "@")
	if len(arr1) != 2 {
		return false
	}

	if len(arr1[0]) == 0 || len(arr1[1]) == 0 {
		return false
	}

	arr2 := strings.Split(arr1[1], ".")
	if len(arr2) != 2 {
		return false
	}

	if len(arr2[0]) == 0 || len(arr2[1]) == 0 {
		return false
	}
	return true
}

func isValidPassword(str string) bool {
	if isEmpty(str) {
		return false
	}

	if len(str) < 5 || len(str) > 30 {
		return false
	}

	capitals := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowers := "abcdefghijklmnopqrstuvwxyz"
	nums := "0123456789"

	if !strings.ContainsAny(str, capitals) || !strings.ContainsAny(str, lowers) || !strings.ContainsAny(str, nums) {
		return false
	}
	return true
}

func isValidUsername(str string) bool {
	if isEmpty(str) {
		return false
	}

	if len(str) < 3 || len(str) > 20 {
		return false
	}

	allowed := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_.-"
	counter := 0

	for _, v := range str {
		if strings.ContainsAny(string(v), allowed) {
			counter++
		}
	}

	if counter != len(str) {
		return false
	}

	return true
}
