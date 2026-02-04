# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Camera Dashboard - A React surveillance dashboard for monitoring 4 HLS camera streams from the "Huajin Nam Định" industrial facility. Displays real-time video feeds with connection status tracking and fullscreen viewing capability.

## Commands

```bash
npm run dev      # Start dev server (port 5173, auto-opens browser)
npm run build    # Production build to dist/
npm run preview  # Preview production build
```

## Architecture

**Tech Stack:** React 18 + Vite 5 + HLS.js

**Component Hierarchy:**
```
App.jsx (state: cameras, cameraStatus, isLive, selectedCamera)
├── Header (title, online/offline counts, back button, clock)
├── CameraGrid (2x2 grid layout, fullscreen mode)
│   └── CameraFrame (HLS stream player, status indicators)
└── Footer (LIVE indicator, control buttons)
```

**Data Flow:**
- Camera config defined in `App.jsx` (`initialCameras` array with stream URLs)
- `CameraFrame` reports connection status up via `onStatusChange` callback
- Click camera → `selectedCamera` state → fullscreen mode
- Each component has co-located CSS file (e.g., `Header/Header.jsx` + `Header/Header.css`)

**HLS Streaming (CameraFrame.jsx):**
- Uses hls.js library with low-latency settings
- Handles MANIFEST_PARSED, FRAG_LOADED, ERROR events
- Auto-reconnect on failure (5 second delay)
- Falls back to native HLS on Safari/iOS

**Styling:**
- CSS variables defined in `src/index.css` (colors, spacing, transitions)
- Dark theme with cyan accents
- Grid layout responsive (1 column on mobile, 2x2 on desktop)

## Stream Configuration

Camera streams configured in `App.jsx`:
```javascript
{
    id: 1,
    name: 'Camera 1',
    location: 'Khu xưởng A',
    isRecording: true,
    streamUrl: 'http://222.252.19.183:54542/drone_stream1/index.m3u8',
}
```

## Language

UI text is in Vietnamese. Date/time formatted with `vi-VN` locale.
