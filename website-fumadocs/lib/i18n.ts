import { defineI18n } from 'fumadocs-core/i18n';

export const i18n = defineI18n({
  defaultLanguage: 'ko',
  languages: ['ko', 'en'],
  // default(ko)는 prefix 없음(/, /docs), en은 /en, /en/docs
  hideLocale: 'default-locale',
});
