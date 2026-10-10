interface TelegramMainButton {
  setText(text: string): void;
  show(): void;
  hide(): void;
  enable(): void;
  disable(): void;
  showProgress(leaveActive?: boolean): void;
  hideProgress(): void;
  onClick(handler: () => void): void;
  offClick(handler: () => void): void;
}

interface TelegramWebApp {
  initData: string;
  ready(): void;
  expand(): void;
  MainButton: TelegramMainButton;
  HapticFeedback: {
    notificationOccurred(type: "error" | "success" | "warning"): void;
    impactOccurred(style: "light" | "medium" | "heavy"): void;
  };
}

interface Window {
  Telegram?: { WebApp: TelegramWebApp };
}
