package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

func main() {
	lambda.Start(Handler)
}

func Handler(event events.S3Event) (string, error) {

	if len(event.Records) < 1 {
		fmt.Println("[Info]no records are provided in the event.")
		return "", errors.New("[Info]no records are provided in the event")
	}

	eventRecord := event.Records[0]
	fmt.Println("[Info]------target event------")
	fmt.Println(eventRecord)
	fmt.Println("[Info]------------------------")
	s3Entity := eventRecord.S3
	regionName := eventRecord.AWSRegion
	bucketName := s3Entity.Bucket.Name
	originalFilename := s3Entity.Object.Key

	rule := ConfigureRules().ChooseRule(originalFilename)
	if rule.Path == "" {
		fmt.Println("[Info]'" + originalFilename + "' does not match with the rules. skipping...")
		return "[Info]'" + originalFilename + "' does not match with the rules. skipping...", nil
	}

	client := newS3Client(regionName)

	//元データ取得
	result, err := getObject(client, bucketName, originalFilename)
	if err != nil {
		return "", err
	}

	image, err := decode(originalFilename, result.Body)
	if err != nil {
		return "", err
	}

	//----------------- 一時ファイルに圧縮データを書き込んでS3に保存 -----------------------------
	tempDirName, err := createTempDir(originalFilename)
	if err != nil {
		return "", err
	}
	for _, outputSpec := range rule.OutputSpecs {
		tempFile, err := os.Create(tempDirName + "/" + calcMD5Hash(originalFilename)) // ファイルが既に存在する場合は消して上書きされる
		if err != nil {
			return "", err
		}
		// NOTE: サイズ調整できる(x,yどちらかを0にすると縦横比そのままでリサイズされる)
		encode(originalFilename, tempFile, resize.Resize(outputSpec.X, outputSpec.Y, image, resize.Bilinear))
		params := &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    formatResizedFilename(originalFilename, outputSpec),
			Body:   tempFile,
		}
		_, err = client.PutObject(params)
		if err != nil {
			tempFile.Close()
			os.RemoveAll(tempDirName)
			return "", err
		}
		tempFile.Close()
	}
	os.RemoveAll(tempDirName)
	//--------------------------------------------------------------------------------------

	fmt.Println("[Info]resized image(s) successfully uploaded to " + bucketName + "!")
	return "[Info]resized image(s) successfully uploaded to " + bucketName + "!", nil
}

/***********************************************************************************/

//TODO: from profile?
func newS3Client(regionName string) *s3.S3 {
	accessKeyId := os.Getenv("TEST_AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")

	sess := session.Must(session.NewSession())
	creds := credentials.NewStaticCredentials(accessKeyId, secretAccessKey, "")
	svc := s3.New(
		sess,
		aws.NewConfig().WithRegion(regionName).WithCredentials(creds),
	)
	return svc
}

func getObject(svc *s3.S3, bucketName string, originalFilename string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(originalFilename),
	}
	result, err := svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result, nil
}

// hogehoge.original.jpeg -> hogehoge.300x300.jpeg
func formatResizedFilename(originalFilename string, spec OutputSpec) *string {
	if spec.Directory != "#ORIG_DIR" {
		return &spec.Directory
	}
	extension := filepath.Ext(originalFilename)
	i := strings.Index(originalFilename, ".original.")
	basename := originalFilename[0:i]
	resizedFilename := basename + "." + strconv.Itoa(int(spec.X)) + "x" + strconv.Itoa(int(spec.Y)) + extension
	return &resizedFilename
}

func decode(filename string, body io.ReadCloser) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg":
		return jpeg.Decode(body)
	case ".jpeg":
		return jpeg.Decode(body)
	case ".png":
		return png.Decode(body)
	case ".gif":
		return gif.Decode(body)
	}
	return nil, errors.New("unexpected file extension '" + ext + "' given.")
}

func encode(filename string, out *os.File, image image.Image) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg":
		jpeg.Encode(out, image, nil)
	case ".jpeg":
		jpeg.Encode(out, image, nil)
	case ".png":
		png.Encode(out, image)
	case ".gif":
		gif.Encode(out, image, nil)
	}
	return errors.New("unexpected file extension '" + ext + "' given.")
}

func createTempDir(text string) (string, error) {
	hashString := calcMD5Hash(text)
	dir := "/tmp/" + hashString
	err := os.Mkdir(dir, 0700)
	return dir, err
}

func calcMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
