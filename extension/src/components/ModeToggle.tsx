import type { SubtitleMode } from '../types';

const MODES: SubtitleMode[] = ['translated', 'bilingual'];

export function ModeToggle({
  mode,
  onChange,
}: {
  mode: SubtitleMode;
  onChange: (mode: SubtitleMode) => void;
}) {
  return (
    <div>
      {MODES.map((value) => (
        <button
          key={value}
          type="button"
          aria-pressed={mode === value}
          onClick={() => onChange(value)}>
          {value === 'translated' ? 'Translated' : 'Bilingual'}
        </button>
      ))}
    </div>
  );
}
