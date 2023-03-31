package s3gateway

import (
	"context"
	"fmt"
	"hitminer-file-manager/util"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func (svr *S3Server) GetObjects(ctx context.Context, filePath, objectName string) error {
	// 如果为根目录,则 prefix = "", 开始不能是 /
	objectName = filepath.ToSlash(objectName)
	objectName = strings.TrimPrefix(objectName, "/")
	// filepath.Clean会删除最后的/
	if strings.HasSuffix(objectName, "/") {
		objectName = filepath.Clean(objectName) + "/"
	} else {
		objectName = filepath.Clean(objectName)
	}
	objectName = filepath.ToSlash(objectName)

	if objectName == "." {
		return fmt.Errorf("illegal path")
	}

	// 判断objectName是一个对象还是文件夹
	originalObjectName := objectName
	if !strings.HasSuffix(objectName, "/") {
		if ok, _ := svr.headObject(ctx, objectName); !ok {
			objectName = objectName + "/"
		}
	}

	// filepath: aa/bb/ object: cc/dd[/..]  -> aa/bb/dd/..
	// filepath: aa/bb/ object: cc/dd/[..]  -> aa/bb/..
	// filepath: aa/bb/ object: cc/dd       -> aa/bb/dd
	// filepath: aa/bb  object: cc/dd[/..]  -> aa/bb/..
	// filepath: aa/bb  object: cc/dd/[..]  -> aa/bb/..
	// filepath: aa/bb  object: cc/dd       -> aa/bb
	if strings.HasSuffix(objectName, "/") {
		for obj := range svr.listObjects(ctx, objectName, "") {
			object := obj
			if !object.IsDirectory && !strings.HasSuffix(object.FullPath, "/") {
				svr.mg.Add()
				go func() {
					defer svr.mg.Done()
					var localPath string
					if strings.HasSuffix(filePath, string(os.PathSeparator)) || filePath == "." {
						if !strings.HasSuffix(originalObjectName, "/") {
							// filepath: aa/bb/ object: cc/dd[/..]  -> aa/bb/dd/..
							p := 0
							if filepath.Dir(originalObjectName) != "." {
								p = len(filepath.Dir(originalObjectName))
							}
							localPath = filepath.Join(filePath, object.FullPath[p:])
						} else {
							// filepath: aa/bb/ object: cc/dd/[..]  -> aa/bb/..
							localPath = filepath.Join(filePath, object.FullPath[len(objectName):])
						}
					} else {
						// filepath: aa/bb  object: cc/dd[/..]  -> aa/bb/..
						// filepath: aa/bb  object: cc/dd/[..]  -> aa/bb/..
						localPath = filepath.Join(filePath, object.FullPath[len(objectName):])
					}
					st, err := os.Stat(localPath)
					if err == nil {
						if st.IsDir() {
							svr.mg.AppendError(fmt.Errorf("fileName: %s is a directory", localPath))
							return
						}
					}
					body, err := svr.getObject(ctx, object.FullPath)
					if err != nil {
						svr.mg.AppendError(err)
						return
					}
					defer func() {
						_ = body.Close()
					}()

					err = os.MkdirAll(filepath.Dir(localPath), 0755)
					if err != nil {
						svr.mg.AppendError(err)
						return
					}

					f, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, 0755)
					if err != nil {
						svr.mg.AppendError(err)
						return
					}
					defer func() {
						_ = f.Close()
					}()

					bar := util.NewBarReader(body, int64(object.Size), "download: "+localPath)
					_, _ = io.Copy(f, bar)
				}()
			}
		}
	} else {
		svr.mg.Add()
		go func() {
			defer svr.mg.Done()
			var localPath string
			if strings.HasSuffix(filePath, string(os.PathSeparator)) || filePath == "." {
				// filepath: aa/bb/ object: cc/dd       -> aa/bb/dd
				p := 0
				if filepath.Dir(originalObjectName) != "." {
					p = len(filepath.Dir(originalObjectName))
				}
				localPath = filepath.Join(filePath, objectName[p:])
			} else {
				// filepath: aa/bb  object: cc/dd       -> aa/bb
				localPath = filepath.Join(filePath, objectName[len(objectName):])
			}
			st, err := os.Stat(localPath)
			if err == nil {
				if st.IsDir() {
					svr.mg.AppendError(fmt.Errorf("fileName: %s is a directory", localPath))
					return
				}
			}
			body, err := svr.getObject(ctx, objectName)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}
			defer func() {
				_ = body.Close()
			}()

			err = os.MkdirAll(filepath.Dir(localPath), 0755)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}

			f, err := os.OpenFile(localPath, os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				svr.mg.AppendError(err)
				return
			}
			defer func() {
				_ = f.Close()
			}()

			_, size := svr.headObject(ctx, objectName)
			bar := util.NewBarReader(body, size, "download: "+localPath)
			_, _ = io.Copy(f, bar)
		}()
	}
	svr.mg.Finish()
	return svr.mg.GetError()
}

func (svr *S3Server) getObject(ctx context.Context, objectName string) (io.ReadCloser, error) {
	reqUrl := &url.URL{
		Scheme: "https",
		Host:   svr.host,
		Path:   "/s3/" + objectName,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+svr.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
