// this checks for the URL to see if it has an UUID as a parameter
function getURLParameter(name) {
	return decodeURIComponent((new RegExp('[?|&]' + name + '=' + '([^&;]+?)(&|#|;|$)').exec(location.search) || [null, ''])[1].replace(/\+/g, '%20')) || null;
}


// simple random name generator
function fetchNames(nameType) {
  let names = [];

  switch (nameType) {
	case 'female':
	  names = ['Berthefried', 'Tatiana', 'Hildeburg', 'Lily', 'Daisy'];
	  break;
	case 'male':
	  names = ['Bilbo', 'Frodo', 'Theodulph', 'Lotho'];
	  break;
	case 'surnames':
	  names = ['Baggins', 'Lightfoot', 'Boulderhill', 'Brockhouse', 'Boffin'];
	  break;
  }

  return { data: names };
}

function pickRandom(list) {
  return list[Math.floor(Math.random() * list.length)];
}

function generateName(gender) {
  // Fetch the names
  const firstNames = fetchNames(gender || pickRandom(['male', 'female']));
  const lastNames = fetchNames('surnames');

  // Pick a random name from each list
  const firstName = pickRandom(firstNames.data);
  const lastName = pickRandom(lastNames.data);

  // Use a template literal to format the full name
  return `${firstName} ${lastName}`;
}