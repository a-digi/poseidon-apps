import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

const en = {
  'app.title': 'Music Hype',
  'settings.title': 'Settings',
  'settings.infoText': "The orchestrator is a background service that automatically syncs chart data. If you don't want data to sync, you can stop it below — it won't restart until you reopen this plugin.",
  'settings.crawlers': 'Plarforms',
  'orchestrator.running': 'Running',
  'orchestrator.stopped': 'Stopped',
  'orchestrator.start': 'Start',
  'orchestrator.stop': 'Stop',
  'orchestrator.restart': 'Restart',
  'orchestrator.pid': 'pid {{pid}}',
  'orchestrator.startedAgo': 'started {{ago}}',
  'orchestrator.heartbeat': 'heartbeat {{ago}}',
  'crawler.runNow': 'Refresh data now',
  'crawler.everyN': 'every {{n}} min',
  'crawler.fileCount_one': '{{count}} item',
  'crawler.fileCount_other': '{{count}} items',
  'crawler.lastData': 'last got data {{ago}}',
  'crawler.never': 'never',
  'items.empty': 'No items yet.',
  'items.showMore': 'Show more',
  'items.loading': 'Loading…',
  'youtube.title': 'YouTube suggestions',
  'youtube.notReady': 'Suggestions not yet available — please check back in a moment.',
  'youtube.tabLabel': 'Suggestion {{n}}',
  'menu.deactivate': 'Deactivate',
  'crawler.active': 'Active',
  'crawler.inactive': 'Inactive',
  'overflow.menu': 'Menu',
};

i18n.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
  resources: {
    en: { translation: en },
  },
});

export default i18n;
