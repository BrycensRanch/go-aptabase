# Go-Aptabase

This project provides a simple and efficient way to integrate your Go applications with [Aptabase](https://aptabase.com), tracking events, sessions, and more! 🌟

---

## 📦 Installation

To get started, simply add this SDK to your Go project with:

```bash
go get github.com/brycensranch/go-aptabase
```

## ❤️‍🔥 Features

- No CGO
- No external dependencies
- Windows, Linux, macOS, and FreeBSD Support
- No random command prompts opening! On Windows, it uses the Windows API as it should!

## 🚀 Usage

Check out the example usage [here](./example/main.go).

To use [Aptabase](https://aptabase.com), you need to provide a API key from the Aptabase dashboard.

## Example data

```json
[
  {
    "eventName": "SIGHUP_EVENT_0",
    "props": {
      "email": "johndosdaae0@example.com",
      "username": "johndodssdsde0"
    },
    "sessionId": "173426498109710887",
    "systemProps": {
      "appBuildNumber": "133",
      "appVersion": "1.0.0",
      "deviceModel": "Predator PH315-54",
      "engineName": "go",
      "engineVersion": "go1.23.4",
      "isDebug": true,
      "locale": "en_US",
      "osName": "Fedora Linux",
      "osVersion": "42",
      "sdkVersion": "go-aptabase@1.0.0"
    },
    "timestamp": "2024-12-15T12:16:21Z"
  },
  {
    "eventName": "SIGHUP_EVENT_1",
    "props": {
      "email": "johndosdaae1@example.com",
      "username": "johndodssdsde1"
    },
    "sessionId": "173426498109710887",
    "systemProps": {
      "appBuildNumber": "133",
      "appVersion": "1.0.0",
      "deviceModel": "Predator PH315-54",
      "engineName": "go",
      "engineVersion": "go1.23.4",
      "isDebug": true,
      "locale": "en_US",
      "osName": "Fedora Linux",
      "osVersion": "42",
      "sdkVersion": "go-aptabase@1.0.0"
    },
    "timestamp": "2024-12-15T12:16:21Z"
  },
  {
    "eventName": "SIGHUP_EVENT_2",
    "props": {
      "email": "johndosdaae2@example.com",
      "username": "johndodssdsde2"
    },
    "sessionId": "173426498109710887",
    "systemProps": {
      "appBuildNumber": "133",
      "appVersion": "1.0.0",
      "deviceModel": "Predator PH315-54",
      "engineName": "go",
      "engineVersion": "go1.23.4",
      "isDebug": true,
      "locale": "en_US",
      "osName": "Fedora Linux",
      "osVersion": "42",
      "sdkVersion": "go-aptabase@1.0.0"
    },
    "timestamp": "2024-12-15T12:16:21Z"
  }
]
```

### 📝 License

This project is licensed under the MIT license like other Aptabase SDKs... If I would usually license it under a GNU license but I don't want to worry any developers. Check out the [LICENSE](./LICENSE) file.
