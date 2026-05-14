import { useEffect, useRef } from 'react';
import Phaser from 'phaser';
import { Connection } from './net/Connection';
import { ArenaScene } from './scenes/ArenaScene';
import type { ArenaSceneInit, InputVector, MobileInputRef } from './scenes/ArenaScene';

interface PhaserGameProps {
  connection: Connection;
  mode: 'desktop' | 'mobile';
}

const ARENA_W = 800;
const ARENA_H = 600;
const JOYSTICK_OUTER_PX = 120;
const JOYSTICK_THUMB_PX = 40;
const JOYSTICK_RADIUS = JOYSTICK_OUTER_PX / 2;
const DEAD_ZONE = 8;

export function PhaserGame({ connection, mode }: PhaserGameProps) {
  const parentRef = useRef<HTMLDivElement | null>(null);
  const gameRef = useRef<Phaser.Game | null>(null);
  const sceneRef = useRef<ArenaScene | null>(null);
  const mobileInputRef = useRef<InputVector>({ up: false, down: false, left: false, right: false });
  const joystickRef = useRef<HTMLDivElement | null>(null);
  const thumbRef = useRef<HTMLDivElement | null>(null);
  const pointerIdRef = useRef<number | null>(null);

  useEffect(() => {
    if (gameRef.current !== null) return;
    const parent = parentRef.current;
    if (parent === null) return;

    const scene = new ArenaScene();
    sceneRef.current = scene;

    const sharedMobileRef: MobileInputRef = mobileInputRef;

    const baseConfig: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: ARENA_W,
      height: ARENA_H,
      backgroundColor: '#eef2f7',
      parent,
      scene,
      banner: false,
    };

    const config: Phaser.Types.Core.GameConfig =
      mode === 'mobile'
        ? {
            ...baseConfig,
            scale: {
              mode: Phaser.Scale.FIT,
              parent,
              autoCenter: Phaser.Scale.CENTER_BOTH,
              width: ARENA_W,
              height: ARENA_H,
            },
          }
        : baseConfig;

    const game = new Phaser.Game(config);
    gameRef.current = game;

    const init: ArenaSceneInit = {
      sendInput: (input) => connection.sendInput(input),
      inputSource: mode === 'mobile' ? 'mobile' : 'keyboard',
      mobileInputRef: sharedMobileRef,
    };
    game.scene.start(ArenaScene.KEY, init);

    return () => {
      sceneRef.current = null;
      gameRef.current = null;
      game.destroy(true);
    };
  }, [connection, mode]);

  useEffect(() => {
    connection.setCallbacks({
      onSnapshot: (msg) => {
        sceneRef.current?.applySnapshot(msg);
      },
      onLeft: (msg) => {
        sceneRef.current?.removePlayer(msg.playerId);
      },
    });
  }, [connection]);

  const setThumb = (dx: number, dy: number) => {
    const thumb = thumbRef.current;
    if (thumb === null) return;
    thumb.style.transform = `translate(${dx}px, ${dy}px)`;
  };

  const updateFromOffset = (dx: number, dy: number) => {
    const dist = Math.hypot(dx, dy);
    const clampedDist = Math.min(dist, JOYSTICK_RADIUS);
    const angle = Math.atan2(dy, dx);
    const cx = Math.cos(angle) * clampedDist;
    const cy = Math.sin(angle) * clampedDist;
    setThumb(cx, cy);

    const v = mobileInputRef.current;
    if (dist < DEAD_ZONE) {
      v.up = false;
      v.down = false;
      v.left = false;
      v.right = false;
      return;
    }
    if (Math.abs(dx) > Math.abs(dy)) {
      v.up = false;
      v.down = false;
      v.left = dx < 0;
      v.right = dx > 0;
    } else {
      v.left = false;
      v.right = false;
      v.up = dy < 0;
      v.down = dy > 0;
    }
  };

  const releaseJoystick = () => {
    pointerIdRef.current = null;
    setThumb(0, 0);
    const v = mobileInputRef.current;
    v.up = false;
    v.down = false;
    v.left = false;
    v.right = false;
  };

  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (pointerIdRef.current !== null) return;
    pointerIdRef.current = e.pointerId;
    e.currentTarget.setPointerCapture(e.pointerId);
    const rect = e.currentTarget.getBoundingClientRect();
    const dx = e.clientX - (rect.left + rect.width / 2);
    const dy = e.clientY - (rect.top + rect.height / 2);
    updateFromOffset(dx, dy);
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (pointerIdRef.current !== e.pointerId) return;
    const rect = e.currentTarget.getBoundingClientRect();
    const dx = e.clientX - (rect.left + rect.width / 2);
    const dy = e.clientY - (rect.top + rect.height / 2);
    updateFromOffset(dx, dy);
  };

  const handlePointerUp = (e: React.PointerEvent<HTMLDivElement>) => {
    if (pointerIdRef.current !== e.pointerId) return;
    try {
      e.currentTarget.releasePointerCapture(e.pointerId);
    } catch {
      /* pointer already released */
    }
    releaseJoystick();
  };

  if (mode === 'mobile') {
    return (
      <div className="relative h-full w-full">
        <div ref={parentRef} className="absolute inset-0" />
        <div
          ref={joystickRef}
          role="application"
          aria-label="Movement joystick"
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onPointerCancel={handlePointerUp}
          className="absolute bottom-6 left-6 flex items-center justify-center rounded-full bg-slate-300/50 touch-none select-none"
          style={{ width: JOYSTICK_OUTER_PX, height: JOYSTICK_OUTER_PX }}
        >
          <div
            ref={thumbRef}
            className="rounded-full bg-slate-700"
            style={{
              width: JOYSTICK_THUMB_PX,
              height: JOYSTICK_THUMB_PX,
              transform: 'translate(0px, 0px)',
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <div
      ref={parentRef}
      className="mx-auto rounded-lg border border-slate-200 bg-slate-100 overflow-hidden"
      style={{ width: ARENA_W, height: ARENA_H }}
    />
  );
}
