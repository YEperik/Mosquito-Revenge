# Mosquito Revenge — Ultimate Arcade Edition

Game 2D top-down stealth arcade menggunakan Go dan Ebitengine v2. Semua karakter, arena, HUD, musuh, telegraph serangan, serta efek partikel digambar saat runtime menggunakan vector/shape drawing. Tidak ada aset gambar eksternal yang diperlukan. Kanvas logis: **360 × 640 portrait**.

## Fitur game

- **5 stage** dengan tema, warna, tekanan angin, target darah dan kombinasi musuh berbeda: Dorm Room, Hot Kitchen, Greenhouse, Night Garden, dan Stormy Rooftop.
- **3 kesulitan** yang dapat dipilih dari menu: Chill (4 nyawa, angin/musuh lebih ringan), Normal (3 nyawa), dan Nightmare (angin, musuh, serta itch meter lebih agresif).
- **Boss event di stage terakhir**: Giant Swatter menandai lokasi serangan, memberi jeda peringatan, lalu menyerang. Hindari zona merah untuk mendapat bonus skor. Ada maksimal tiga pola serangan dalam stage final.
- **3 jenis musuh**: Spider berpatroli, Dragonfly mengejar pemain, dan Beetle melakukan charge.
- **3 skill dengan cooldown**: DASH untuk kabur cepat, CLOAK agar aman dari deteksi sementara, dan PULSE untuk menahan musuh.
- **4 power-up**: Nectar (skor dan progres), Cool Mist (menurunkan itch), Shield (menahan satu benturan), dan Wing Overdrive (meningkatkan kecepatan serta dash).
- Combo saat menghisap darah, perut membesar dan memerah, efek partikel, floating feedback, damage flash, peringatan tepukan dan kipas dengan gaya angin yang berubah per stage.
- **Peringkat 1–3 medali per stage** berdasarkan jumlah benturan, power-up dan perfect escape.
- **Pause/resume**, restart run, layar game-over dan victory dengan ringkasan statistik.
- Kontrol multi-touch: satu jari mengarahkan nyamuk, sementara jari lain dapat menekan skill.
- CI melakukan gofmt check, unit tests, desktop compile, WASM build dan APK build.

## Kontrol

| Aksi | Android/touch | PC |
| --- | --- | --- |
| Mulai / lanjut | Tap layar | Enter atau klik kiri |
| Pilih kesulitan | Tap CHILL/NORMAL/NIGHTMARE pada menu | Klik pilihan pada menu |
| Terbang | Drag jari | WASD atau tombol panah |
| Menghisap darah | Tahan di zona tangan | Tahan Space atau klik kiri sambil berada di atas tangan |
| DASH | Tombol DASH di kanan bawah | Shift kiri/kanan |
| CLOAK | Tombol CLOAK di kiri bawah | E |
| PULSE | Tombol PULSE di tengah bawah | Q |
| Pause/resume | Tombol `II` di kanan atas, tap layar saat pause untuk resume | P atau Esc |
| Restart saat pause | — | R |

## Versi Go dan menjalankan desktop

Persyaratan: **Go 1.25.x atau lebih baru**, toolchain native untuk platform desktop, dan akses internet pada build pertama kali.

```sh
go mod download
go test ./...
go run .
```

Di Linux, Ebitengine perlu paket pengembangan untuk OpenGL/X11/ALSA. Pada Ubuntu/Debian:

```sh
sudo apt-get update
sudo apt-get install -y libasound2-dev libgl1-mesa-dev libx11-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```

## WebAssembly lokal

```sh
go run github.com/hajimehoshi/wasmserve@latest .
```

Atau build situs manual:

```sh
mkdir -p site
GOOS=js GOARCH=wasm go build -trimpath -o site/mosquito-revenge.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/wasm_exec.js
cp web/index.html site/index.html
```

Sajikan folder `site` melalui HTTP, bukan `file://`.

## Build APK Android lokal

Persyaratan: Go 1.25+, JDK 17, Android SDK Platform 35, Build Tools 35.0.0, Android NDK `27.2.12479018`, dan Gradle `8.11.1`.

```sh
mkdir -p mobile/android/app/libs

go run github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile bind \
  -target android \
  -androidapi 23 \
  -javapkg com.example.mosquitorevenge \
  -o mobile/android/app/libs/mosquitorevenge.aar \
  ./mobile

gradle --no-daemon -p mobile/android assembleDebug
```

APK debug dihasilkan ke `mobile/android/app/build/outputs/apk/debug/app-debug.apk`. Perintah `ebitenmobile bind` membuat AAR; Gradle mengemasnya bersama Android Activity menjadi APK.

## GitHub Actions dan Pages

Workflow `/.github/workflows/build.yml` menjalankan tests/build terlebih dahulu, lalu membangun WASM dan APK secara paralel.

1. Push folder proyek ke repository GitHub, branch `main` atau `master`.
2. Buka **Actions** dan lihat job `Go checks and gameplay tests`, `Build WebAssembly`, dan `Build Android debug APK`.
3. Unduh artefak `mosquito-revenge-android-debug` dari run yang berhasil.
4. Untuk menerbitkan web, buka **Settings → Pages → Build and deployment**, lalu pilih **GitHub Actions**. Push ke branch utama akan memicu deployment.

APK yang disediakan adalah **debug APK**, bukan APK bertanda tangan rilis Play Store. Distribusi produksi memerlukan signing config dan keystore yang disimpan sebagai GitHub Actions Secrets, bukan di repository.

## Struktur proyek

```text
mosquito-revenge/
├── .github/workflows/build.yml
├── game/
│   ├── game.go          # state, difficulty, gameplay, boss, input, power-up
│   ├── draw.go          # vector rendering, efek visual, HUD dan overlay
│   └── game_test.go     # unit tests gameplay dasar
├── mobile/
│   ├── mobile.go
│   └── android/         # Gradle, manifest dan Activity
├── web/index.html
├── go.mod
├── main.go
└── README.md
```

## Catatan validasi

Jalankan `go test ./...` dan workflow GitHub Actions setelah mengunduh. Lingkungan penyusun saat ini tidak memiliki Go 1.25+ beserta dependency cache dan koneksi unduhan Go dari container dibatasi, sehingga saya tidak menyatakan APK/WASM berhasil dibangun di sini. Workflow menyediakan verifikasi otomatis pada runner Ubuntu dengan toolchain yang sesuai.
