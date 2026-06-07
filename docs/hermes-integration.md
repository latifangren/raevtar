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
- Protected claim endpoint: `POST /api/v1/editorial-inbox/claim`
- Protected detail endpoint: `GET /api/v1/editorial-inbox/{itemID}`
- Protected update endpoint: `POST /api/v1/editorial-inbox/{itemID}`
- Protected complete endpoint: `POST /api/v1/editorial-inbox/{itemID}/complete`
- Protected fail endpoint: `POST /api/v1/editorial-inbox/{itemID}/fail`

Semua endpoint di atas pakai **admin key** seperti protected API lain di Raevtar.

---

## Protected Project Content API

Di luar editorial inbox, Raevtar sekarang juga punya project content flow yang bisa dikelola operator atau agent lewat protected API.

Endpoint yang tersedia:

- `GET /api/v1/projects`
- `GET /api/v1/projects/{slug}/updates`
- `GET /api/v1/projects/{slug}/changelog`
- `GET /api/v1/projects/{slug}/relations`
- `GET /api/v1/projects/{slug}/showcase`
- `POST /api/v1/projects`
- `PUT /api/v1/projects/{projectID}`
- `DELETE /api/v1/projects/{projectID}`
- `POST /api/v1/projects/{projectID}/updates`
- `PUT /api/v1/projects/updates/{updateID}`
- `DELETE /api/v1/projects/updates/{updateID}`
- `POST /api/v1/projects/{projectID}/relations`
- `DELETE /api/v1/projects/relations/{relationID}`
- `POST /api/v1/projects/{projectID}/showcase`
- `PUT /api/v1/projects/showcase/{itemID}`
- `DELETE /api/v1/projects/showcase/{itemID}`

Catatan penting:

- `GET /api/v1/projects` bersifat public-safe dan hanya mengembalikan project yang published
- query yang didukung untuk list public:
  - `featured=true`
  - `state=planning|active|paused|shipped|archived`
  - `sort=newest|oldest`
- read endpoint child resources (`updates`, `changelog`, `relations`, `showcase`) juga public-safe dan hanya mengembalikan surface yang sudah published / public-facing
- endpoint write (`POST`/`PUT`/`DELETE`) tetap **protected** dan butuh `Authorization: Bearer <RAEVTAR_ADMIN_KEY>`
- update/delete pakai **numeric project ID**, bukan slug
- slug project dihasilkan saat create dan dipertahankan saat update

Ini berguna kalau nanti Hermes atau operator workflow mau:

1. publish artikel biasa ke `/api/v1/posts`
2. atau publish build log / showcase project ke `/api/v1/projects`

Dengan begitu, blog dan project archive tetap terpisah secara sengaja.

### Payload create/update project

```json
{
  "title": "Whyred Watchtower",
  "content_md": "# Whyred Watchtower\n\nProject build log...",
  "excerpt": "Ringkasan singkat buat archive card.",
  "cover_image_url": "/uploads/watchtower.png",
  "published": true,
  "state": "active",
  "featured": true,
  "sort_order": 1,
  "tags": ["oss", "self-hosted"]
}
```

Semantics tambahan:

- `title` dan `content_md` wajib ada
- `state` menentukan lifecycle publik project (`planning`, `active`, `paused`, `shipped`, `archived`)
- `featured` menentukan apakah project masuk featured lane publik
- `sort_order` menentukan urutan internal di antara project featured/public
- `sort_order < 0` akan dinormalisasi server menjadi `0`
- response create/update mengembalikan object project lengkap, termasuk `id`, `slug`, dan tags yang sudah dinormalisasi

### Payload build log / changelog project update

```json
{
  "kind": "build_log",
  "title": "Collector tuning pass",
  "content_md": "## Notes\n\n- Reduced noise\n- Improved queue ordering",
  "published": true,
  "pinned": false,
  "sort_order": 0,
  "event_at": "2026-06-07T13:30:00Z"
}
```

Catatan:

- `kind` yang valid: `timeline`, `build_log`, `changelog`
- updates ini yang kemudian muncul di project detail timeline dan `/projects/{slug}/changelog`
- cocok untuk Hermes kalau outputnya bukan satu page overwrite penuh, tapi tambahan progres / release note ke project yang sudah ada

### Payload relation dan showcase

Relation ke post/project lain:

```json
{
  "target_type": "post",
  "target_id": 12,
  "relation_kind": "related",
  "sort_order": 0
}
```

Showcase item terstruktur:

```json
{
  "kind": "image",
  "title": "Operator overview",
  "body_md": "Screenshot atau artefak yang menjelaskan state terbaru.",
  "asset_url": "/uploads/watchtower-overview.png",
  "external_url": "",
  "embed_provider": "",
  "embed_ref": "",
  "published": true,
  "sort_order": 0
}
```

### Kapan pakai posts vs projects

Gunakan `posts` kalau outputnya adalah artikel blog kronologis yang masuk archive kategori biasa.

Gunakan `projects` kalau outputnya adalah:

- build log yang ingin hidup di `/projects`
- showcase project yang perlu featured lane di homepage/public archive
- artefak yang lebih cocok diposisikan sebagai public project entry daripada dispatch blog

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

### Intake operator yang dipakai sekarang

Di admin UI, operator **tidak perlu lagi** mengisi semua field teknis di atas saat create item baru.

Flow intake yang sekarang sengaja dipersingkat jadi:

- `source_value` — repo/link/tema/project source
- `category_hint` — opsional
- `note` — opsional, buat angle/judul proyek/catatan editor

Sisanya diisi default oleh server untuk flow admin biasa:

- `source_type = repo`
- `priority = 50`
- `not_before = now`
- `mode = scheduled_assignment`
- `status = approved`

Artinya: kalau operator masukin repo atau tema lalu menjalankan tick Hermes, item itu **langsung eligible** buat diambil tanpa langkah approval tambahan.

### Mode yang tersedia

- `scheduled_assignment`
- `opportunistic_assignment`
- `campaign_theme`
- `autonomous_seed`

### Status yang tersedia

- `queued`
- `approved`
- `paused`
- `running`
- `failed`
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

### Step 2 — Claim candidate inbox secara aman

Hermes claim satu item yang ready:

```text
POST /api/v1/editorial-inbox/claim
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
Content-Type: application/json
```

Payload:

```json
{
  "worker": "hermes-cron"
}
```

Kalau hasilnya `204 No Content`:

- Hermes masuk mode autonomous seperti flow yang sekarang

Kalau hasilnya ada:

- item langsung masuk status `running`
- response berisi `item` + `claim_token`
- claim ini punya lease TTL server-side 30 menit
- **tidak boleh ada 2 item `running` aktif sekaligus**

Kalau masih ada item lain dengan lease `running` yang masih aktif:

- claim baru akan diblok
- response tetap `204 No Content`
- Hermes tidak boleh mulai ngerjain item kedua sampai item pertama selesai, gagal, atau lease-nya kadaluarsa

Ini sengaja dibuat supaya satu tick / satu worker tidak bikin dua publish flow jalan paralel tanpa kontrol.

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

### Step 5 — Mark item selesai / gagal

Kalau publish sukses, Hermes update item inbox:

```text
POST /api/v1/editorial-inbox/{itemID}/complete
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
Content-Type: application/json
```

Payload:

```json
{
  "claim_token": "<opaque-token>",
  "published_post_id": 123
}
```

Kalau publish gagal, Hermes lapor gagal secara eksplisit:

```text
POST /api/v1/editorial-inbox/{itemID}/fail
Authorization: Bearer <RAEVTAR_ADMIN_KEY>
Content-Type: application/json
```

Payload retryable:

```json
{
  "claim_token": "<opaque-token>",
  "failure_note": "publish API 500",
  "failure_meta": "{\"status\":500}",
  "retryable": true
}
```

Semantics:

- `retryable: true` → item kembali ke `approved` dengan backoff server-side via `not_before`
- `retryable: false` → item pindah ke `failed`
- stale lease dipulihkan otomatis setelah expiry
- item yang pernah dieksekusi akan mempertahankan `attempt_count`
- item yang sudah pernah di-claim tidak lagi dianggap mutable dari admin intake flow biasa

Backoff yang dipakai sekarang:

- attempt 1 → `15m`
- attempt 2 → `1h`
- attempt 3+ → `6h`

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
  "source_value": "https://github.com/zed-industries/zed",
  "category_hint": "tools",
  "note": "Bahas angle editor cepat buat workflow developer Linux."
}
```

Untuk **admin UI**, payload mental model yang benar memang sesederhana itu. Server akan mengisi default field internal.

Untuk **API machine-to-machine**, payload eksplisit penuh tetap boleh dipakai kalau memang butuh kontrol detail.

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

Contract aktual sekarang juga mengekspose lease policy seperti:

- `lease_ttl = 30m`
- retry schedule
- fairness policy
- overdue priority
- **single-running gate**: claim baru diblok selama masih ada `running` item dengan lease aktif

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

### 2. Item yang belum pernah dieksekusi masih boleh diubah atau dihapus

Flow yang sekarang dipakai:

- item dengan `attempt_count == 0` masih boleh **edit** dan **delete**
- begitu item pernah di-claim sekali saja, item dianggap **locked after first execution attempt**

Jadi delete memang ada, tapi cuma aman dipakai sebelum Hermes mulai kerja.

### 3. Jangan pakai edit/delete untuk item yang sudah pernah dijalankan

Kalau item sudah pernah dieksekusi, perlakukan lifecycle-nya lewat status/runtime flow:

- `done`
- `cancelled`
- `failed`

daripada mencoba “reset manual” item dari admin intake. Audit trail dan attempt history jadi tetap utuh.

### 4. Gunakan `not_before` untuk fleksibilitas waktu bila memang perlu

Kalau kamu mau item dikerjakan di slot tertentu, tidak perlu hardcode slot Hermes di admin.
Cukup set:

- `not_before`
- `priority`
- optional `deadline`

Hermes tetap akan menyesuaikan dengan cadence cron yang sudah ada.

### 5. Jangan overload satu tick dengan banyak item

Untuk fase awal, satu tick sebaiknya fokus ke satu keputusan publish. Ini bikin debugging dan observability jauh lebih gampang.

Dan itu sekarang memang enforced juga oleh control plane:

- satu item `running` aktif dulu
- item berikutnya baru bisa diambil setelah yang current selesai/fail/stale

---

## Status Implementasi Saat Ini

Yang **sudah ada** sekarang:

- source of truth editorial inbox
- admin UI untuk operator
- protected machine-readable contract
- protected API create/list/detail/update/claim/complete/fail/summary
- ready-ordering semantics di service
- Hermes-facing claim lease TTL
- retry backoff server-side
- fairness policy antara queue lane dan autonomous lane
- overdue-first priority
- publish analytics summary
- single active running item guard
- admin-side edit/delete guard untuk item yang belum pernah dieksekusi
- simplified admin intake flow dengan default server-side

Yang **belum** tetap masuk area lanjutan:

- Hermes-side implementation yang benar-benar mengonsumsi contract ini secara otomatis
- batch merge beberapa item jadi satu artikel
- anti-repeat repo/category policy lintas run
- richer editorial heuristics di sisi Hermes prompt/runtime

---

## Fase Lanjutan yang Disarankan

Kalau integrasi awal ini sudah berjalan, langkah berikutnya yang paling masuk akal:

### Phase 2 — Hermes consumption runtime refinement

Sekarang claim contract dasarnya sudah ada. Fase berikutnya lebih ke menyambungkan Hermes runtime agar:

- claim item saat tick jalan
- fallback ke autonomous saat `204 No Content`
- complete/fail item secara konsisten

### Phase 3 — Richer editorial policy

Tambahkan policy seperti:

- overdue escalation
- anti-repeat repo/category
- optional fairness antara queue lane dan autonomous lane

---

## Ringkasan Singkat

Kalau mau diringkas jadi satu kalimat:

> Hermes tetap jadi penulis otonom, tapi sebelum bergerak sendiri dia harus cek meja editor Raevtar dulu.

Itu inti integrasi ini.
