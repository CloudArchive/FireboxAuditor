---
name: project-context
description: Firebox Auditor proje kuralları. Her kod değişikliğinde uygulanır.
---

# Firebox Auditor — Proje Kuralları

## Stack
- Backend: Go 1.23, Gin framework, JWT (golang-jwt/v5), bcrypt
- Frontend: React + Vite + Tailwind CSS
- Storage: Dosya tabanlı (data/ klasörü, JSON)
- i18n: TR/EN JSON dosyaları — `frontend/src/i18n/tr.json` ve `en.json`

## i18n Kuralı (KRİTİK)
- Kullanıcıya görünen HER metin `t('key')` ile çekilmeli
- Hardcoded Türkçe veya İngilizce string yasak
- Yeni key eklendiğinde HER ZAMAN hem tr.json hem en.json güncellenmeli
- Key yapısı: `section.subKey` (örn: `auth.loginBtn`, `dashboard.noAudits`)

## Brand Kuralları
- Renkler CSS variable ile: `var(--wg-red)`, `var(--wg-blue)`, `var(--wg-headline)`
- Tailwind class'larında: `text-wg-red`, `bg-wg-red`, `border-wg-gray-light`
- Primary action butonları: `bg-wg-red hover:bg-wg-red-hover text-white`
- Yeni renk tanımlama — `frontend/src/index.css` içindeki `:root` bloğuna ekle

## Kod Kalitesi
- Go: her public fonksiyona yorum satırı
- React: prop drilling yerine context kullan (AuthContext, I18nContext)
- Hata mesajları kullanıcıya gösterilen yerler için i18n key kullan
- API fetch işlemleri AuthContext'teki `apiFetch()` ile yapılmalı (token otomatik eklenir)

## Dosya Yapısı
- Go backend: tek klasör (root)
- React pages: `frontend/src/pages/`
- React components: `frontend/src/components/`
- React contexts: `frontend/src/contexts/`
