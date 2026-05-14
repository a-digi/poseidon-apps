import Phaser from 'phaser';
import type { Snapshot } from '../types';

export interface InputVector {
  up: boolean;
  down: boolean;
  left: boolean;
  right: boolean;
}

export interface InputPayload extends InputVector {
  seq: number;
}

export interface MobileInputRef {
  current: InputVector;
}

export interface ArenaSceneInit {
  sendInput: (input: InputPayload) => void;
  inputSource: 'keyboard' | 'mobile';
  mobileInputRef: MobileInputRef;
}

interface PlayerEntry {
  circle: Phaser.GameObjects.Arc;
  label: Phaser.GameObjects.Text;
  target: { x: number; y: number };
}

const ARENA_W = 800;
const ARENA_H = 600;
const PLAYER_RADIUS = 16;
const LERP = 0.2;
const LABEL_OFFSET_Y = -20;

export class ArenaScene extends Phaser.Scene {
  static readonly KEY = 'ArenaScene';

  private readonly players = new Map<string, PlayerEntry>();
  private readonly snapshotIds = new Set<string>();
  private readonly lastInput: InputVector = { up: false, down: false, left: false, right: false };
  private readonly currentInput: InputVector = { up: false, down: false, left: false, right: false };
  private seq = 0;
  private sendInput: (input: InputPayload) => void = () => {};
  private inputSource: 'keyboard' | 'mobile' = 'keyboard';
  private mobileInputRef: MobileInputRef = { current: { up: false, down: false, left: false, right: false } };
  private cursors: Phaser.Types.Input.Keyboard.CursorKeys | null = null;
  private keys: { W: Phaser.Input.Keyboard.Key; A: Phaser.Input.Keyboard.Key; S: Phaser.Input.Keyboard.Key; D: Phaser.Input.Keyboard.Key } | null = null;

  constructor() {
    super({ key: ArenaScene.KEY });
  }

  init(data: ArenaSceneInit): void {
    this.sendInput = data.sendInput;
    this.inputSource = data.inputSource;
    this.mobileInputRef = data.mobileInputRef;
  }

  create(): void {
    this.add.rectangle(ARENA_W / 2, ARENA_H / 2, ARENA_W, ARENA_H, 0xeef2f7).setOrigin(0.5);
    const border = this.add.rectangle(ARENA_W / 2, ARENA_H / 2, ARENA_W, ARENA_H);
    border.setStrokeStyle(1, 0xe2e8f0);
    border.setFillStyle(undefined, 0);

    if (this.inputSource === 'keyboard' && this.input.keyboard !== null) {
      this.cursors = this.input.keyboard.createCursorKeys();
      this.keys = {
        W: this.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.W),
        A: this.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.A),
        S: this.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.S),
        D: this.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.D),
      };
    }
  }

  applySnapshot(snap: Snapshot): void {
    this.snapshotIds.clear();
    for (const p of snap.players) {
      this.snapshotIds.add(p.id);
      const entry = this.players.get(p.id);
      if (entry === undefined) {
        const color = Phaser.Display.Color.HexStringToColor(p.color).color;
        const circle = this.add.arc(p.x, p.y, PLAYER_RADIUS, 0, 360, false, color);
        circle.setOrigin(0.5);
        const label = this.add.text(p.x, p.y + LABEL_OFFSET_Y, p.name, {
          fontFamily: "'Inter', sans-serif",
          fontSize: '12px',
          color: '#ffffff',
          stroke: '#000000',
          strokeThickness: 2,
        });
        label.setOrigin(0.5);
        this.players.set(p.id, { circle, label, target: { x: p.x, y: p.y } });
      } else {
        entry.target.x = p.x;
        entry.target.y = p.y;
      }
    }

    for (const [id, entry] of this.players) {
      if (!this.snapshotIds.has(id)) {
        entry.circle.destroy();
        entry.label.destroy();
        this.players.delete(id);
      }
    }
  }

  removePlayer(id: string): void {
    const entry = this.players.get(id);
    if (entry === undefined) return;
    entry.circle.destroy();
    entry.label.destroy();
    this.players.delete(id);
  }

  update(): void {
    for (const entry of this.players.values()) {
      entry.circle.x += (entry.target.x - entry.circle.x) * LERP;
      entry.circle.y += (entry.target.y - entry.circle.y) * LERP;
      entry.label.x = entry.circle.x;
      entry.label.y = entry.circle.y + LABEL_OFFSET_Y;
    }

    this.readInputInto(this.currentInput);
    if (
      this.currentInput.up !== this.lastInput.up ||
      this.currentInput.down !== this.lastInput.down ||
      this.currentInput.left !== this.lastInput.left ||
      this.currentInput.right !== this.lastInput.right
    ) {
      this.lastInput.up = this.currentInput.up;
      this.lastInput.down = this.currentInput.down;
      this.lastInput.left = this.currentInput.left;
      this.lastInput.right = this.currentInput.right;
      this.seq += 1;
      this.sendInput({
        up: this.currentInput.up,
        down: this.currentInput.down,
        left: this.currentInput.left,
        right: this.currentInput.right,
        seq: this.seq,
      });
    }
  }

  private readInputInto(out: InputVector): void {
    if (this.inputSource === 'mobile') {
      const src = this.mobileInputRef.current;
      out.up = src.up;
      out.down = src.down;
      out.left = src.left;
      out.right = src.right;
      return;
    }

    const c = this.cursors;
    const k = this.keys;
    out.up = (c?.up.isDown ?? false) || (k?.W.isDown ?? false);
    out.down = (c?.down.isDown ?? false) || (k?.S.isDown ?? false);
    out.left = (c?.left.isDown ?? false) || (k?.A.isDown ?? false);
    out.right = (c?.right.isDown ?? false) || (k?.D.isDown ?? false);
  }
}
