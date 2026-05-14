import { useEffect, useState } from 'react';

interface FlipNumberProps {
  value: number;
  className?: string;
}

export function FlipNumber({ value, className }: FlipNumberProps) {
  const digits = String(value).padStart(2, '0').split('');
  return (
    <div className={className}>
      {digits.map((digit, i) => (
        <FlipDigit key={i} digit={digit} />
      ))}
    </div>
  );
}

function FlipDigit({ digit }: { digit: string }) {
  const [displayed, setDisplayed] = useState(digit);
  const [flipKey, setFlipKey] = useState(0);

  useEffect(() => {
    if (digit === displayed) return;
    setDisplayed(digit);
    setFlipKey(k => k + 1);
  }, [digit, displayed]);

  return (
    <span key={flipKey} className="flip-tile font-black tabular-nums leading-none select-none">
      {displayed}
    </span>
  );
}
