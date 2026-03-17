---
name: add-audit-rule
description: Yeni bir WatchGuard audit kuralı ekler (audit.go + i18n)
---

# Yeni Audit Kuralı Ekleme Workflow'u

Aşağıdaki adımları sırayla uygula:

## 1. audit.go'ya kural fonksiyonu ekle
```go
// Rule N (Severity): Kısa açıklama
func checkXxx(cfg *WatchGuardConfig) AuditResult {
    r := AuditResult{
        RuleID:   "RULE-00N",
        Severity: High, // Critical | High | Medium
        Passed:   true,
    }
    // kontrol mantığı
    if len(r.Details) > 0 {
        r.Passed = false
    }
    return r
}
```

## 2. RunAudit() fonksiyonuna ekle
`audit.go` içindeki `results` slice'ına yeni fonksiyonu ekle.

## 3. i18n key'lerini ekle (HER İKİ DOSYAYA)
`frontend/src/i18n/tr.json` ve `en.json` içindeki `rules` objesine:
```json
"RULE-00N": {
  "name": "...",
  "descPass": "...",
  "descFail": "... {{details}} ...",
  "remediation": "..."
}
```

## 4. Test
`audit_test.go` içine test case ekle.
