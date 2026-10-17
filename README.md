# Mosquito Revenge — Ultimate Arcade Edition

Game 2D top-down stealth arcade menggunakan Go dan Ebitengine v2. Kanvas logis **360 × 640 portrait**. Semua karakter, arena, HUD, musuh, telegraph serangan, dan efek partikel digambar saat runtime menggunakan geometri Ebitengine; game tidak membutuhkan sprite atau gambar eksternal.

## Fitur game

- **5 stage** dengan tema, warna, tekanan angin, target darah, dan komposisi musuh berbeda.
- **3 tingkat kesulitan**: Chill, Normal, dan Nightmare.
- **Boss Giant Swatter** pada stage terakhir, dengan serangan bertanda area, waktu peringatan, dan bonus saat berhasil menghindar.
- **3 tipe musuh**: Spider berpatroli, Dragonfly mengejar pemain, dan Beetle menerjang.
- **3 skill**: DASH, CLOAK, dan PULSE, semuanya memiliki cooldown.
- **4 power-up**: Nectar, Cool Mist, Shield, dan Wing Overdrive.
- Combo menghisap darah, skor, meter gatal, perut yang membesar/merah, efek partikel, feedback damage, perfect escape, serta kipas yang mendorong gerakan.
- Sistem nyawa, peringkat stage 1–3 bintang, pause/resume, layar game over dan victory, statistik run.
- Kontrol mouse/keyboard dan multi-touch: satu jari mengarahkan nyamuk, jari lain dapat menekan skill.
- GitHub Actions memeriksa file wajib, menjalankan tes, build desktop, build WASM, build APK debug, dan deploy GitHub Pages pada branch utama.

## Struktur proyek

```text
mosquito-revenge/
├── .github/workflows/build.yml
├── game/
│   ├── game.go
│   ├── draw.go
│   └── game_test.go
├── mobile/
│   ├── mobile.go
│   └── android/
│       ├── settings.gradle
│       ├── build.gradle
│       ├── gradle.properties
│       └── app/
│           └── src/main/ (manifest, Activity Java, styles.xml)
├── web/index.html
├── go.mod
├── main.go
└── README.md
```

## Persyaratan

- **Go 1.25.x** (sesuai directive di `go.mod`).
- Desktop Linux: toolchain native serta paket OpenGL/X11/ALSA.
- Android: JDK 17, Android SDK Platform 35, Build Tools 35.0.0, NDK `27.2.12479018`, Gradle 8.11.1.

## Menjalankan desktop

Jalankan dari root proyek:

```sh
go mod download
go test ./...
go run .
```

Di Ubuntu/Debian, dependensi desktop Ebitengine bisa dipasang dengan:

```sh
sudo apt-get update
sudo apt-get install -y libasound2-dev libgl1-mesa-dev libx11-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxrender-dev \
  libxxf86vm-dev
```

## Menjalankan web lokal

Cara termudah untuk pengembangan:

```sh
go run github.com/hajimehoshi/wasmserve@latest .
```

Atau susun file WASM untuk di-host:

```sh
mkdir -p site
GOOS=js GOARCH=wasm go build -trimpath -o site/mosquito-revenge.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/wasm_exec.js
cp web/index.html site/index.html
```

Sajikan folder `site` melalui HTTP, bukan `file://`. File `wasm_exec.js` harus diambil dari toolchain Go yang sama dengan yang digunakan untuk mengompilasi WASM.

## Build APK Android lokal

Siapkan Go 1.25+, JDK 17, Android SDK Platform 35, Build Tools 35.0.0, NDK `27.2.12479018`, dan Gradle 8.11.1. Pastikan `ANDROID_HOME` atau `ANDROID_SDK_ROOT` menunjuk SDK Android yang benar.

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

`ebitenmobile bind` menghasilkan **AAR**; Gradle mengemas AAR bersama Activity Android menjadi APK. APK debug dihasilkan di:

```text
mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

APK debug belum ditandatangani untuk rilis produksi Play Store.

## GitHub Actions dan Pages

1. Unggah **seluruh isi folder proyek** ke repository GitHub, bukan hanya `main.go` dan folder `game`.
2. Pastikan path ini ada di repository: `web/index.html`, seluruh file `mobile/android/`, `mobile/mobile.go`, dan `game/game_test.go`. Workflow akan menggagalkan job lebih awal dengan pesan jelas jika berkas wajib hilang.
3. Push ke `main` atau `master`, lalu buka tab **Actions**.
4. Jika job Android lulus, unduh artefak `mosquito-revenge-android-debug` dari workflow run.
5. Untuk publikasi web, buka **Settings → Pages → Build and deployment**, lalu pilih **GitHub Actions**. Deploy otomatis dijalankan saat push ke `main` atau `master`.

### Memeriksa kegagalan build

- **`Required project file is missing or empty`**: repository tidak memuat semua isi arsip. Pastikan file yang dilaporkan benar-benar ter-commit.
- **`go.mod requires go >= 1.25.0`**: jangan gunakan Go 1.23/1.24; workflow mengatur Go `1.25.x`.
- **NDK tidak ditemukan**: cek log `Install Android SDK and NDK`. Workflow memasang NDK `27.2.12479018` dan meneruskan `ANDROID_NDK_HOME`/`ANDROID_NDK_ROOT` ke step berikutnya.
- **`EbitenView` tidak ditemukan saat kompilasi Java**: periksa bahwa step pembuatan AAR berhasil dan `-javapkg com.example.mosquitorevenge` tetap cocok dengan import Activity `com.example.mosquitorevenge.mobile.EbitenView`.
- **GitHub Pages tidak ter-deploy**: ubah Pages ke source **GitHub Actions**; job deploy hanya berjalan untuk push ke `main`/`master`, bukan pull request.

Workflow dan konfigurasi proyek sudah diaudit secara statis, tetapi keberhasilan build end-to-end tetap harus dikonfirmasi di runner GitHub atau komputer yang memiliki akses internet dan toolchain lengkap. Lingkungan pembuatan berkas ini tidak dapat mengunduh toolchain Go 1.25 dari `proxy.golang.org`, jadi saya tidak mengklaim APK/WASM telah berhasil dikompilasi di sini.

## Kontrol

| Aksi | Android / touch | PC |
|---|---|---|
| Mulai / lanjut / retry | Tap layar | Enter atau klik kiri |
| Pilih tingkat kesulitan | Tap Chill / Normal / Nightmare | Klik pilihan |
| Terbang | Drag jari | WASD / tombol panah |
| Menghisap darah | Tahan di atas tangan | Tahan Space atau klik kiri di zona tangan |
| DASH | Tombol DASH | Shift |
| CLOAK | Tombol CLOAK | E |
| PULSE | Tombol PULSE | Q |
| Pause / resume | Tombol II, tap overlay | P atau Esc |
| Restart saat pause | — | R |
