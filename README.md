# Mosquito Revenge — Ultimate Arcade Edition

A 2D top-down stealth arcade game built using Go and Ebitengine v2. It features a logical canvas of **360 × 640 (portrait orientation)**. All characters, arenas, HUD elements, enemies, attack telegraphs, and particle effects are rendered at runtime using Ebitengine’s geometry primitives; the game requires no external sprites or image assets.

## Game Features

- **5 stages**, each with unique themes, color palettes, wind pressure, blood-consumption targets, and enemy compositions.
- **3 difficulty levels**: Chill, Normal, and Nightmare.
- A **Giant Swatter boss** in the final stage, featuring area-of-effect attacks, warning indicators, and bonuses for successful dodges.
- **3 enemy types**: Spiders (patrolling), Dragonflies (pursuing the player), and Beetles (charging).
- **3 skills**: DASH, CLOAK, and PULSE—all with individual cooldowns.
- **4 power-ups**: Nectar, Cool Mist, Shield, and Wing Overdrive.
- Gameplay mechanics including blood-sucking combos, a scoring system, an "itch" meter, a visually expanding/reddening abdomen, particle effects, damage feedback, "perfect escape" bonuses, and fans that influence movement.
- A life system, 1–3 star stage ratings, pause/resume functionality, game-over and victory screens, and run statistics.
- Support for mouse/keyboard and multi-touch controls (e.g., one finger to steer the mosquito, while others activate skills).
- GitHub Actions integration for mandatory file checks, automated testing, desktop and WASM builds, debug APK generation, and deployment to GitHub Pages from the main branch. ## Project structure

```text
mosquito-revenge/
├── .github/workflows/build.yml
├── games/
│ ├── game.go
│ ├── draw.go
│ └── game_test.go
├── mobile/
│ ├── mobile.go
│ └── android/
│ ├── settings. gradle
│ ├── build.gradle
│ ├── gradle. properties
│ └── app/
│ └── src/main/ (manifest, Activity Java, styles.xml)
├── web/index.html
├── go. mod
├── main.go
└── README.md
```
## Requirements

* **Go 1.25.x** (as specified in the `go.mod` directive).
* Linux Desktop: native toolchain and OpenGL/X11/ALSA packages.
* Android: JDK 17, Android SDK Platform 35, Build Tools 35.0.0, NDK `27.2.12479018`, and Gradle 8.11.1.

## Running on desktop

Execute from the project root:

```sh
go mod download
go test ./...
go run .
```

On Ubuntu/Debian, Ebitengine's desktop dependencies can be installed via:

```sh
sudo apt-get update
sudo apt-get install -y libasound2-dev libgl1-mesa-dev libx11-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxrender-dev \
  libxxf86vm-dev
```
## Running a local web server

The easiest way for development:

```sh
go run github.com/hajimehoshi/wasmserve@latest .
```

Alternatively, build the WASM files for hosting:

```sh
mkdir -p site
GOOS=js GOARCH=wasm go build -trimpath -o site/mosquito-revenge.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/wasm_exec.js
cp web/index.html site/index.html
```

Serve the `site` folder via HTTP, not `file://`. The `wasm_exec.js` file must be retrieved from the same Go toolchain used to compile the WASM.

## Building an Android APK locally

Set up Go 1.25+, JDK 17, Android SDK Platform 35, Build Tools 35.0.0, NDK `27.2.12479018`, and Gradle 8.11.1. Ensure that `ANDROID_HOME` or `ANDROID_SDK_ROOT` points to the correct Android SDK.

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

`ebitenmobile bind` generates an **AAR**; Gradle packages the AAR, along with an Android Activity, into an APK. The debug APK is generated at:

```text
mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

This debug APK is not signed for a production release on the Play Store.

## GitHub Actions and Pages

1. Upload the **entire contents of the project folder** to the GitHub repository—not just `main.go` and the `game` folder.
2. Ensure the following paths exist in the repository: `web/index.html`, all files within `mobile/android/`, `mobile/mobile.go`, and `game/game_test.go`. The workflow will fail the job early and provide a clear error message if any required files are missing.
3. Push your changes to the `main` or `master` branch, then navigate to the **Actions** tab.
4. If the Android job passes, download the `mosquito-revenge-android-debug` artifact from the workflow run.
5. For web deployment, go to **Settings → Pages → Build and deployment** and select **GitHub Actions**. Deployment is triggered automatically upon pushing to the `main` or `master` branch.

### Troubleshooting build failures

* **`Required project file is missing or empty`**: The repository does not contain the full contents of the project archive. Verify that the specified files have been properly committed.
* **`go.mod requires go >= 1.25.0`**: Do not use Go 1.23 or 1.24; the workflow is configured for Go `1.25.x`.
* **NDK not found**: Check the logs for the `Install Android SDK and NDK` step. The workflow installs NDK version `27.2.12479018` and passes the `ANDROID_NDK_HOME`/`ANDROID_NDK_ROOT` variables to subsequent steps.
* **`EbitenView` not found during Java compilation**: Ensure that the AAR generation step completed successfully and that the `-javapkg com.example.mosquitorevenge` parameter matches the Activity import: `com.example.mosquitorevenge.mobile.EbitenView`.
* **GitHub Pages not deploying**: Ensure that the source for GitHub Pages is set to **GitHub Actions**. Note that the deployment job only runs when pushing to the `main` or `master` branch—it does not trigger for pull requests. The project workflow and configuration have undergone static analysis; however, a successful end-to-end build must still be verified on a GitHub runner or a machine equipped with internet access and the complete toolchain. This build environment is unable to download the Go 1.25 toolchain from `proxy.golang.org`, so I cannot confirm that the APK/WASM files have been successfully compiled here.

## Controls

| Action | Android / Touch | PC |
|---|---|---|
| Start / Continue / Retry | Tap the screen | Enter or left-click |
| Select difficulty level | Tap Chill / Normal / Nightmare | Click your choice |
| Fly | Drag with finger | WASD / arrow keys |
| Drain blood | Hold over the hand | Hold Space or left-click on the hand zone |
| DASH | DASH button | Shift |
| CLOAK | CLOAK button | E |
| PULSE | PULSE button | Q |
| Pause / Resume | II button, tap the overlay | P or Esc |
| Restart while paused | — | R |
