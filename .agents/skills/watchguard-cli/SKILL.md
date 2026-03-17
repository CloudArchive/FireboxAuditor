---
name: watchguard-cli
description: WatchGuard Firebox CLI komutları hakkında sorularda veya SSH komutları yazılırken kullan.
---

# WatchGuard Firebox CLI Referansı

## SSH Bağlantı Detayları
- Default port: 4118 (standart SSH 22 değil)
- Protokol: Interactive CLI shell (exec-style çalışmaz, PTY gerekir)
- Prompt: `WG#` (admin) veya `WG>` (read-only)

## Kritik Komutlar
| Komut | Çıktı |
|---|---|
| `sysinfo` | Model, SN, Uptime, CPU, Memory |
| `show feature-key` | Lisans ve servis bilgileri |
| `show certificate` | Sertifika listesi ve expiry |
| `show signature-update` | GAV/IPS imza güncelleme tarihleri |
| `show connection count` | Aktif bağlantı sayısı |
| `show vpn-status bovpn gateway` | VPN tünel durumu |
| `who` | Bağlı kullanıcılar |

## SSH Output Temizleme
Prompt satırları (`WG#`, `WG>`) çıktıdan çıkarılmalı.
ANSI control character'lar temizlenmeli.
Fonksiyon: `stripSSHNoise()` — `handlers.go` içinde mevcut.

## Önemli Notlar
- `session.Run()` Firebox'ta ÇALIŞMAZ — interactive PTY gerekir
- `export config to console` komutu desteklenmeyebilir (model bağımlı)
- Feature key parse: "Feature:" ve "Expiration:" satırları aranır
