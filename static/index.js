function getExp() {
  const foundationDate = new Date('2025-06-27');
  const today = new Date();

  let diffTime = today - foundationDate;
  let diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

  if (diffDays < 0) {
    diffDays = 0;
  }

  // Если больше или равно 365, считаем годы
  if (diffDays >= 365) {
    let years = Math.floor(diffDays / 365);

    const lastDigit = years % 10;
    const lastTwoDigits = years % 100;

    let suffix;

    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
      suffix = "років";
    } else {
      switch (lastDigit) {
        case 1:
          suffix = "рік";
          break;
        case 2:
        case 3:
        case 4:
          suffix = "роки";
          break;
        default:
          suffix = "років";
      }
    }

    document.getElementById("exp").textContent = years + " " + suffix;
  } else {
    // Иначе считаем дни, как раньше
    const lastDigit = diffDays % 10;
    const lastTwoDigits = diffDays % 100;

    let suffix;

    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
      suffix = "днів";
    } else {
      switch (lastDigit) {
        case 1:
          suffix = "день";
          break;
        case 2:
        case 3:
        case 4:
          suffix = "дні";
          break;
        default:
          suffix = "днів";
      }
    }

    document.getElementById("exp").textContent = diffDays + " " + suffix;
  }
}

document.addEventListener("DOMContentLoaded", () => {
  getExp();
});