# pariksha 🧪

[![Go Report Card](https://goreportcard.com/badge/github.com/DevNavix/pariksha)](https://goreportcard.com/report/github.com/DevNavix/pariksha)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/DevNavix/pariksha/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.20+-blue)](https://golang.org/dl/)

**pariksha** is a lightweight and expressive testing + benchmarking toolkit for Go, tailored specifically to HTTP API test cases using the [Gin](https://github.com/gin-gonic/gin) web framework.

Inspired by the Sanskrit word _**Pariksha**_ (परीक्षा), meaning **"test"** or **"examination"**, this framework is built to help developers write **cleaner, faster, and more structured** test code with minimal boilerplate.

---

## ✨ Features

- ✅ Declarative API test case structure
- 🔁 Reusable setup for HTTP requests and `gin.Context`
- 🧪 Native support for `*testing.T` and `*testing.B`
- ⚙️ Integration with Go’s `httptest` package
- 🚀 Designed for speed, clarity, and code reuse

---

## 📦 Installation

```bash
go get github.com/DevNavix/pariksha@v1.0.1
