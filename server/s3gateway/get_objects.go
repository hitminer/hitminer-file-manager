package s3gateway

import (
	"context"
	"fmt"
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
		objectName = ""
	}

	// 判断objectName是一个对象还是文件夹
	originalObjectName := objectName
	if !strings.HasSuffix(objectName, "/") && objectName != "" {
		if ok, _ := svr.HeadObject(ctx, objectName); !ok {
			objectName = objectName + "/"
		}
	}
	if filePath == "." {
		filePath, _ = os.Getwd()
	}

	fileCnt, fileTotalSize := int64(0), int64(0)
	for obj := range svr.ListObjects(ctx, objectName, "") {
		if !obj.IsDirectory && !strings.HasSuffix(obj.FullPath, "/") {
			fileCnt++
			fileTotalSize += int64(obj.Size)
		}
	}
	if fileCnt > 128 && (fileTotalSize/fileCnt) <= 32*1024*1024 {
		svr.bar.NewCntBar(fileCnt, "download files")
		svr.bar.SetPrint(false)
	}

	// filepath: aa/bb/ Object: cc/dd[/..]  -> aa/bb/dd/..
	// filepath: aa/bb/ Object: cc/dd/[..]  -> aa/bb/..
	// filepath: aa/bb/ Object: cc/dd       -> aa/bb/dd
	// filepath: aa/bb  Object: cc/dd[/..]  -> aa/bb/..
	// filepath: aa/bb  Object: cc/dd/[..]  -> aa/bb/..
	// filepath: aa/bb  Object: cc/dd       -> aa/bb
	if strings.HasSuffix(objectName, "/") || objectName == "" {
		for obj := range svr.ListObjects(ctx, objectName, "") {
			object := obj
			if !object.IsDirectory && !strings.HasSuffix(object.FullPath, "/") {
				svr.mg.Add()
				go func() {
					defer svr.mg.Done()
					defer func() {
						_, _ = svr.bar.Write(nil)
					}()
					var localPath string
					if strings.HasSuffix(filePath, string(os.PathSeparator)) || filePath == "." {
						if !strings.HasSuffix(originalObjectName, "/") && originalObjectName != "" {
							// filepath: aa/bb/ Object: cc/dd[/..]  -> aa/bb/dd/..
							p := 0
							if filepath.Dir(originalObjectName) != "." {
								p = len(filepath.Dir(originalObjectName))
							}
							localPath = filepath.Join(filePath, object.FullPath[p:])
						} else {
							// filepath: aa/bb/ Object: cc/dd/[..]  -> aa/bb/..
							localPath = filepath.Join(filePath, object.FullPath[len(objectName):])
						}
					} else {
						// filepath: aa/bb  Object: cc/dd[/..]  -> aa/bb/..
						// filepath: aa/bb  Object: cc/dd/[..]  -> aa/bb/..
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

					rel, _ := filepath.Rel(filePath, localPath)
					if rel == "" {
						rel = filepath.Base(localPath)
					}
					bar := svr.bar.NewBarReader(body, int64(object.Size), fmt.Sprintf("download: %s", rel))
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
				// filepath: aa/bb/ Object: cc/dd       -> aa/bb/dd
				p := 0
				if filepath.Dir(originalObjectName) != "." {
					p = len(filepath.Dir(originalObjectName))
				}
				localPath = filepath.Join(filePath, objectName[p:])
			} else {
				// filepath: aa/bb  Object: cc/dd       -> aa/bb
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

			_, size := svr.HeadObject(ctx, objectName)

			rel, _ := filepath.Rel(filePath, localPath)
			if rel == "" {
				rel = filepath.Base(localPath)
			}
			bar := svr.bar.NewBarReader(body, size, fmt.Sprintf("download: %s", rel))
			_, _ = io.Copy(f, bar)
		}()
	}

	svr.mg.Finish()
	svr.bar.Wait()
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
