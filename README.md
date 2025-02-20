# woody_ear
Ухо Вуди - приложение для распознавание аудио-файла и его анализирования посредством AI.
Приложение поддерживает файлы формата .wav и .mp3 (.mp3 в ходе выполнения конвертируется в .wav посредством утилиты sox)
Распознавание аудиофайла происходит посредством библиотеки vosk-api с использованием локальной модели https://alphacephei.com/vosk/models

## Сборка
### Локальная сборка MacOS
Установить sox
```
brew install sox
```

Установить переменные окружения:
```
DYLD_LIBRARY_PATH - путь к директории с файлом libvosk.dylib
CGO_CPPFLAGS - путь к директории с файлом для компиляции vosk_api.h
CGO_LDFLAGS - путь к директории с файлом libvosk.dylib
MODEL_PATH - путь к директории с моделью
```

# woody_ear
Woody Ear is an application for the recognition and analysis of audio files using AI.
The application supports .wav and .mp3 formats. (.mp3 files are converted to .wav during execution using the sox utility.)
Audio file recognition is performed using the vosk-api library with a local model from https://alphacephei.com/vosk/models.

## Build
### Local Build on MacOS
Install sox:
```sh
brew install sox
```

Set environment variables:
- `DYLD_LIBRARY_PATH`: Path to the directory containing the `libvosk.dylib` file
- `CGO_CPPFLAGS`: Path to the directory with the `vosk_api.h` file for compilation
- `CGO_LDFLAGS`: Path to the directory containing the `libvosk.dylib` file
- `MODEL_PATH`: Path to the directory with the model