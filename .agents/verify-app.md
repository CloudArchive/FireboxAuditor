---
name: verify-app
description: Uygulamayı başlatır ve temel akışı test eder
---

# Uygulama Doğrulama

## Adımlar

1. Backend'i başlat: `go run .`
2. Tarayıcıda `http://localhost:8443` aç
3. Login sayfasının geldiğini doğrula
4. `admin` / `admin` ile giriş yap
5. Dashboard'un açıldığını doğrula
6. "Yeni XML Yükle" butonuna tıkla
7. `testdata/sample-config.xml` dosyasını yükle
8. Audit sonuç sayfasının açıldığını doğrula
9. Mavi "SSH ile Zenginleştir" banner'ının göründüğünü doğrula
10. Tüm adımların ekran görüntüsünü al ve walkthrough oluştur

## Beklenen Sonuçlar
- Login: auth.loginTitle görünmeli
- Dashboard: dashboard.recentAudits görünmeli  
- Audit: device.identitySection, liveSection (blurred), licenseSection (blurred) görünmeli
- Score gauge: sample-config için < 50 olmalı