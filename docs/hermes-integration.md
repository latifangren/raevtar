# Hermes Integration — Raevtar Editorial Inbox

Dokumen ini menjelaskan cara menghubungkan **Hermes cron job yang sudah ada** dengan **editorial inbox** di Raevtar.

Scope dokumen ini sengaja dibatasi ke sisi **Raevtar contract + integration flow**.

- **Tidak** mengubah konfigurasi Hermes di mesin remote
- **Tidak** mengubah `jobs.json` Hermes secara langsung
- **Tidak** memaksa Hermes jadi worker statik

Tujuannya: Hermes tetap agentic, tapi selalu **cek meja editor Raevtar dulu** sebelum jalan pakai inisiatif sendiri.

---

## Tujuan

Arsitektur yang dipakai:

- **Raevtar** = editorial control plane / source of truth
- **Hermes cron** = trigger cadence
- **Hermes agent** = reasoning + execution plane

Artinya:

1. Admin/operator mengelola niat editorial dari Raevtar
2. Hermes job yang sudah jalan terjadwal tetap dipakai
3. Saat tick cron datang, Hermes cek inbox dulu
4. Kalau ada item yang eligible, Hermes kerjakan
5. Kalau tidak ada, Hermes kembali ke mode autonomous

---

## Prinsip Integrasi

### 1. Jangan simpan kecerdasan editorial di `jobs.json`

`jobs.json` Hermes cukup jadi scheduler/pemicu.

Logika seperti:

- mana item yang due
- mana yang lebih prioritas
- mana yang boleh diproses sekarang
- status item editorial

harus hidup di **Raevtar**, bukan di file scheduler Hermes.

### 2. Queue bukan rel statik

Editorial inbox **bukan** FIFO worker queue biasa.

Inbox adalah kumpulan **intent editorial** yang nanti dibaca Hermes. Jadi Hermes tetap boleh:

- memilih eksekusi terbaik dari item yang eligible
- menolak angle yang jelek
- fallback ke autonomous mode saat inbox kosong

### 3. Satu tick = satu keputusan publish

Untuk integrasi awal, paling aman menganggap satu run cron Hermes menghasilkan **satu keputusan publish**:

- ambil satu item inbox, atau
- jalan autonomous kalau inbox kosong

Jangan mulai dengan multi-item batching di sisi runtime. Itu bisa ditambahkan belakangan kalau memang dibutuhkan.

---

## Editorial Inbox yang Sudah Ada di Raevtar

Phase 1 control plane sekarang hidup di:

- Admin UI: `/admin/editorial-inbox`
- Protected API contract: `GET /api/v1/editorial-inbox/contract`
- Protected list endpoint: `GET /api/v1/editorial-inbox`
- Protected create endpoint: `POST /api/v1/editorial-inbox`
- Protected detail endpoint: `GET /api/v1/editorial-inbox/{itemID}`
- Protected update endpoint: `POST /api/v1/editorial-inbox/{itemID}`

Semua endpoint di atas pakai **admin key** seperti protected API lain di Raevtar.

---

## Data Model Ringkas

Inbox item menyimpan field berikut:

| Field | Keterangan |
|------|------------|
| `source_type` | jenis sumber, mis. `repo`, `url`, `topic`, `idea` |
| `source_value` | nilai sumber, mis. URL repo atau topik teks |
| `category_hint` | slug kategori blog yang diinginkan, opsional |
| `priority` | prioritas numerik `0..100` |
| `not_before` | item tidak boleh diproses sebelum waktu ini |
| `deadline` | deadline opsional |
| `note` | catatan editorial/operator |
| `mode` | mode niat editorial |
| `status` | status lifecycle item |

### Mode yang tersedia

- `scheduled_assignment`
- `opportunistic_assignment`
- `campaign_theme`
- `autonomous_seed`

### Status yang tersedia

- `queued`
- `approved`
- `paused`
- `done`
- `cancelled`

---

## Ready Queue Semantics

Untuk integrasi awal, Hermes tidak perlu menebak item mana yang harus dipilih.

Raevtar sudah punya semantics “ready” berikut:

- status harus `approved`
- `not_before <= now`
- urutan prioritas:
  1. `priority` tertinggi dulu
  2. `deadline` paling awal dulu, `NULL` di belakang
  3. `created_at` paling lama dulu

Ini penting, karena berarti prompt Hermes tidak perlu menyimpan kebijakan penjadwalan rumit. Dia cukup membaca state dari Raevtar.

---

## Flow Integrasi yang Direkomendasikan

Gunakan alur berikut di job Hermes yang sudah ada.

### Step 1 — Cek contract

Hermes bisa membaca:

```text
GET /api/v1/editorial-inbox/contract
```

Tujuannya bukan untuk tiap run selalu wajib memanggil contract, tapi untuk memastikan sisi Hermes tahu:

- field apa saja yang tersedia
- mode apa saja yang valid
- status apa saja yang valid
- bagaimana Raevtar mendefinisikan item “ready”

### Step 2 — Ambil candidate inbox

Hermes ambil list item yang ready:

```text
GET /api/v1/editorial-inbox?ready=true
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
```

Kalau hasilnya kosong:

- Hermes masuk mode autonomous seperti flow yang sekarang

Kalau hasilnya ada:

- ambil item pertama sebagai kandidat utama
- optionally baca detail item by ID kalau butuh konteks tambahan

### Step 3 — Generate article dari item inbox

Hermes menulis artikel berdasarkan:

- `source_type`
- `source_value`
- `category_hint`
- `note`
- kebijakan editorial job/prompt yang sudah kamu punya

### Step 4 — Publish ke flow blog biasa

Publish tetap lewat endpoint yang sekarang:

```text
POST /api/v1/posts
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
```

### Step 5 — Mark item selesai

Kalau publish sukses, Hermes update item inbox:

```text
POST /api/v1/editorial-inbox/{itemID}
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
Content-Type: application/json
```

Payload update bisa mempertahankan field lama dan hanya mengganti status ke `done`.

Kalau kamu belum mau bikin Hermes mengirim payload update penuh, workaround paling sederhana nanti adalah:

- Hermes baca detail item dulu
- Hermes kirim ulang field existing + `status: done`

---

## Fallback Autonomous Behavior

Kalau inbox kosong, Hermes **tetap jalan seperti sekarang**.

Ini bukan failure condition.

Justru ini desain yang diinginkan:

- ada queue → hormati explicit editorial intent
- queue kosong → Hermes tetap produktif dan inisiatif

Jadi flow yang diinginkan adalah:

```text
tick cron
  -> cek ready inbox
    -> if exists: process inbox item
    -> else: autonomous article generation
```

---

## Contoh Payload

### Buat item inbox baru

```json
{
  "source_type": "repo",
  "source_value": "https://github.com/zed-industries/zed",
  "category_hint": "tools",
  "priority": 80,
  "not_before": "2026-06-06T08:00:00Z",
  "deadline": "2026-06-07T23:00:00Z",
  "note": "Bahas angle editor cepat buat workflow developer Linux.",
  "mode": "scheduled_assignment",
  "status": "approved"
}
```

### Contoh response contract

Contoh shape ringkas:

```json
{
  "source_of_truth": "raevtar",
  "resource": "editorial_inbox",
  "modes": [
    "scheduled_assignment",
    "opportunistic_assignment",
    "campaign_theme",
    "autonomous_seed"
  ],
  "statuses": [
    "queued",
    "approved",
    "paused",
    "done",
    "cancelled"
  ],
  "ready_selection": {
    "status": "approved",
    "not_before_lte": "now",
    "order": [
      "priority desc",
      "deadline asc nulls last",
      "created_at asc"
    ]
  }
}
```

### Contoh update item ke `done`

```json
{
  "source_type": "repo",
  "source_value": "https://github.com/zed-industries/zed",
  "category_hint": "tools",
  "priority": 80,
  "not_before": "2026-06-06T08:00:00Z",
  "deadline": "2026-06-07T23:00:00Z",
  "note": "Bahas angle editor cepat buat workflow developer Linux.",
  "mode": "scheduled_assignment",
  "status": "done"
}
```

---

## Rekomendasi Adaptasi Prompt Hermes

Karena kamu sendiri yang nanti akan menyesuaikan prompt/job Hermes, rekomendasinya adalah menambah blok logika seperti ini di prompt operasionalnya:

1. **Selalu cek editorial inbox dulu**
2. Kalau ada item ready, kerjakan item itu
3. Gunakan `source_type`, `source_value`, `category_hint`, dan `note` sebagai arahan utama
4. Kalau publish sukses, update inbox item ke `done`
5. Kalau inbox kosong, lanjut autonomous generation seperti biasa

Yang penting: jangan ubah Hermes jadi mesin yang cuma menghabiskan queue. Tetap biarkan dia punya ruang judgement terhadap angle, framing, dan kualitas tulisan.

---

## Rekomendasi Policy Operasional

Walau implementasi Hermes-side belum disentuh, dari sekarang lebih aman kalau kamu pakai policy berikut:

### 1. `queued` belum berarti boleh diproses

Gunakan `approved` untuk item yang memang siap dikonsumsi Hermes.

Ini bikin admin panel lebih enak:

- `queued` = ide masuk, belum dilepas
- `approved` = siap dibaca Hermes

### 2. Jangan hapus item, terminalkan statusnya

Lebih baik pakai:

- `done`
- `cancelled`

daripada delete. Audit trail jadi tetap utuh.

### 3. Gunakan `not_before` untuk fleksibilitas waktu

Kalau kamu mau item dikerjakan di slot tertentu, tidak perlu hardcode slot Hermes di admin.
Cukup set:

- `not_before`
- `priority`
- optional `deadline`

Hermes tetap akan menyesuaikan dengan cadence cron yang sudah ada.

### 4. Jangan overload satu tick dengan banyak item

Untuk fase awal, satu tick sebaiknya fokus ke satu keputusan publish. Ini bikin debugging dan observability jauh lebih gampang.

---

## Batasan Phase 1 Saat Ini

Yang **sudah ada** sekarang:

- source of truth editorial inbox
- admin UI untuk operator
- protected machine-readable contract
- protected API create/list/detail/update
- ready-ordering semantics di service

Yang **belum** dikerjakan di fase ini:

- Hermes otomatis claim/lease item
- status `running` / `failed` / retry backoff
- batch merge beberapa item jadi satu artikel
- fairness policy antara queue dan autonomous lane
- overdue escalation logic
- auto-link item inbox ke post ID hasil publish

Semua itu bisa ditambahkan nanti, tapi tidak perlu untuk integrasi awal.

---

## Fase Lanjutan yang Disarankan

Kalau integrasi awal ini sudah berjalan, langkah berikutnya yang paling masuk akal:

### Phase 2 — Hermes consumption contract refinement

Tambahkan flow yang lebih nyaman buat Hermes, misalnya endpoint khusus seperti:

- `GET /api/v1/editorial-inbox?ready=true&limit=1`

atau nantinya endpoint claim khusus jika memang diperlukan.

### Phase 3 — Execution lifecycle

Tambahkan status seperti:

- `running`
- `failed`
- `published`

plus metadata hasil run.

### Phase 4 — Editorial intelligence

Tambahkan policy seperti:

- overdue escalation
- anti-repeat repo/category
- optional fairness antara queue lane dan autonomous lane

---

## Ringkasan Singkat

Kalau mau diringkas jadi satu kalimat:

> Hermes tetap jadi penulis otonom, tapi sebelum bergerak sendiri dia harus cek meja editor Raevtar dulu.

Itu inti integrasi ini.
