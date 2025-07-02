function getTimeOfDay() {
  const now = new Date();
  const hour = now.getHours();

  if (hour >= 5 && hour <= 12) {
    return "Доброго ранку";
  } else if (hour >= 13 && hour <= 18) {
    return "Доброго дня";
  } else if (hour >= 19 && hour <= 21) {
    return "Доброго вечора";
  } else if (hour >= 22 || hour <= 4) {
    return "Доброї ночі";
  }
}

document.addEventListener("DOMContentLoaded", () => {
  const timeLabel = document.getElementById("time");
  timeLabel.textContent = getTimeOfDay();

  const account = document.getElementById("ac");
  const name = document.getElementById("name").textContent;
  const popUp = document.getElementById("account")
  account.addEventListener("click", () => {
    popUp.style.display = popUp.style.display === 'block' ? 'none' : 'block'
  })
  account.textContent = name[0];

});
