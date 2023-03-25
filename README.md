## root

```
hitminer-file-manager client, it can be used to manage hitminer file systems remotely

Usage:
  hitminer-file-manager [flags]
  hitminer-file-manager [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         Get files to hitminer file systems recursively
  help        Help about any command
  login       Login hitminer
  logout      Logout hitminer
  ls          List files in hitminer file systems
  mkdir       Make directory to hitminer file systems
  put         Put files to hitminer file systems recursively
  rm          Remove files to hitminer file systems

Flags:
  -h, --help      help for hitminer-file-manager
  -v, --version   version for hitminer-file-manager
```

## login
```
Login hitminer with your username and password

Usage:
  hitminer-file-manager login [flags]

Flags:
  -h, --help              help for login
      --host string       colony host (default "www.hitminer.cn")
  -p, --password string   password
  -u, --username string   user name

```

## logout
```
Logout hitminer

Usage:
  hitminer-file-manager logout [flags]

Flags:
  -h, --help   help for logout
```

## ls
```
List files in hitminer file systems.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".

Usage:
  hitminer-file-manager ls [remote_path] [flags]

Flags:
  -h, --help   help for ls
```

## rm
```
Remove files to hitminer file systems.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".

Usage:
  hitminer-file-manager rm [remote_path] [flags]

Flags:
  -h, --help        help for rm
  -r, --recursive   remove files recursively
```

## put
```
Put files to hitminer file systems recursively.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".

Usage:
  hitminer-file-manager put [local_path] [remote_path] [flags]

Flags:
  -h, --help   help for put
```

## get
```
Get files to hitminer file systems recursively.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".

Usage:
  hitminer-file-manager get [remote_path] [local_path]  [flags]

Flags:
  -h, --help   help for get
```

## mkdir
```
Make directory to hitminer file systems.
Please note that The directory of the project data has a prefix "project/{project name}/",
The directory of the dataset has a prefix "dataset/".

Usage:
  hitminer-file-manager mkdir [remote_path] [flags]

Flags:
  -h, --help   help for mkdir
```

## upgrade
```
Upgrade hitminer file manager to the latest versions

Usage:
  hitminer-file-manager upgrade [flags]

Flags:
  -h, --help   help for upgrade
```