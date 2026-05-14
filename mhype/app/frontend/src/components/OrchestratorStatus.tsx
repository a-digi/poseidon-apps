import { useTranslation } from 'react-i18next';
import type { OrchestratorStatus as OrchestratorStatusType } from '../api';
import { relativeTime } from '../lib/relativeTime';
import { Spinner } from './Spinner';

type BusySlot = 'start' | 'stop' | 'restart' | null;

interface OrchestratorStatusProps {
  status: OrchestratorStatusType | null;
  onStart: () => void;
  onStop: () => void;
  onRestart: () => void;
  busy: BusySlot;
}

const buttonBase =
  'inline-flex items-center justify-center gap-1.5 px-3 py-1.5 rounded text-sm font-medium ' +
  'disabled:opacity-50 disabled:cursor-not-allowed transition-colors';

export function OrchestratorStatus({ status, onStart, onStop, onRestart, busy }: OrchestratorStatusProps) {
  const { t } = useTranslation();
  const running = !!status?.running;
  const anyBusy = busy !== null;

  return (
    <div className="flex items-center gap-4 flex-wrap">
      <div className="flex flex-col gap-1">
        <span
          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold w-fit ${
            running ? 'bg-green-100 text-green-800' : 'bg-slate-200 text-slate-700'
          }`}
        >
          <span
            className={`w-2 h-2 rounded-full mr-1.5 ${running ? 'bg-green-500' : 'bg-slate-500'}`}
          />
          {running ? t('orchestrator.running') : t('orchestrator.stopped')}
        </span>
        {running && status && (
          <div className="text-xs text-slate-500 flex flex-wrap gap-x-3">
            {status.pid !== undefined && <span>{t('orchestrator.pid', { pid: status.pid })}</span>}
            {status.startedAt && (
              <span>{t('orchestrator.startedAgo', { ago: relativeTime(status.startedAt) })}</span>
            )}
          </div>
        )}
      </div>

      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onStart}
          disabled={running || anyBusy}
          className={`${buttonBase} bg-green-600 hover:bg-green-700 text-white`}
        >
          {busy === 'start' && <Spinner className="w-4 h-4" />}
          {t('orchestrator.start')}
        </button>
        <button
          type="button"
          onClick={onStop}
          disabled={!running || anyBusy}
          className={`${buttonBase} bg-slate-600 hover:bg-slate-700 text-white`}
        >
          {busy === 'stop' && <Spinner className="w-4 h-4" />}
          {t('orchestrator.stop')}
        </button>
        <button
          type="button"
          onClick={onRestart}
          disabled={!running || anyBusy}
          className={`${buttonBase} bg-blue-600 hover:bg-blue-700 text-white`}
        >
          {busy === 'restart' && <Spinner className="w-4 h-4" />}
          {t('orchestrator.restart')}
        </button>
      </div>
    </div>
  );
}
