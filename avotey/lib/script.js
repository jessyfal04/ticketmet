// Requetes AJAX

// Compte
function login() {
	let email = $("#email").val();
	let password = $("#password").val();

	$.ajax({
			method: "GET",
			url: "ajax/compte/login.php",
			data: {
				"email": email,
				"password": password
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				$("#txt-connexion-uuid").html(e.data.uuid);
				toggleConnected(true);

				setMessages("success", e.message);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
});
}

function register() {
	let email = $("#email").val();
	let password = $("#password").val();

	$.ajax({
			method: "GET",
			url: "ajax/compte/register.php",
			data: {
				"email": email,
				"password": password
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				$("#txt-connexion-uuid").html(e.data.uuid);
				toggleConnected(true);
                                setMessages("success", i18n("registered_success"));
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

function logout() {
	$.ajax({
			method: "GET",
			url: "ajax/compte/logout.php",
			data: {},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				toggleConnected(false);
				setMessages("success", e.message);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// Vérifier si l'email existe déjà
function emailExists() {
	let email = $("#email").val();

	$.ajax({
			method: "GET",
			url: "ajax/compte/emailExists.php",
			data: {
				"email": email
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				$("#registerButton").prop("disabled", e.data == 0 ? false : true);
				$("#loginButton").prop("disabled", e.data == 0 ? true : false);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// Change l'affichage en fonction de si l'utilisateur est connecté ou non
function toggleConnected(connected) {
	$(".if-login").css("display", connected ? "initial" : "none");
	$(".if-not-login").css("display", connected ? "none" : "initial");

	clearForms();
	seeScrutins();
}

// SCRUTIN
// Changer de panneau
function seeScrutins(which) {
	// Afficher les blocs
	$("#scrutins").children().hide();
	$("#" + which).show();

	// Boutons
	$("#scrutinsButtons").children().addClass("is-outlined");
	$("#button-"+which).removeClass("is-outlined");

	// button-voterScrutin and resultatsScrutin, if which => display:initial;
	$("#button-voterScrutin").css("display", which == "voterScrutin" ? "initial" : "none");
	$("#button-resultatsScrutin").css("display", which == "resultatsScrutin" ? "initial" : "none");
	

	if (which == "consulterScrutins")
		consulter();

	else if (which == "gererListes")
		getLists();

	else if (which == "createScrutin")
		getVotersLists();
}

// Créer un scrutin
function create() {
	let encrypt = new JSEncrypt();
	let publicKey = encrypt.getPublicKey();
	let privateKey = encrypt.getPrivateKey();

	let question = $("#question").val();
	let systemeVote = $("#systemeVote").val();

	// Get all choices
	let choices = [];
	$("#choicesList input").each(function () {
		choices.push($(this).val());
	});
	choices = JSON.stringify(choices);

	// Get voters emails and voters procurations
	let voters = [];
	$("#votersList input").each(function () {
		if ($(this).attr("type") == "email")
			voters.push({"email": $(this).val(),});
		else 
			voters[voters.length - 1].procuration = $(this).val();
	});
	voters = JSON.stringify(voters);

	$.ajax({
			method: "GET",
			url: "ajax/scrutin/createScrutin.php",
			data: {
				"question": question,
				"systemeVote": systemeVote,
				"choices": choices,
				"voters": voters,
				"publicKey": publicKey
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				seeScrutins("consulterScrutins");
                                setMessages("warning", i18n("ballot_created"));
				clearForms();

				localStorage.setItem(`scrutin-${e.data.scrutinId}`, privateKey);

				sendNotification("ouverture", e.data.scrutinId);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// 		Ajouts rapides
let votersLists = {};
let choicesLists = {
	"Binaire": ["Oui", "Non"],
};

function getVotersLists() {
	$.ajax({
		method: "GET",
		url: "ajax/liste/getLists.php",
		data: {},
		dataType: "json"
	})
	.done(function (e) {
		if (e.status == "success") {
			votersLists = e.data.lists;

			$("#listChoices").html("");
			$("#listChoices").append(`<option disabled selected value>...</option>`);
			for (let [key, value] of Object.entries(votersLists)) {
				$("#listChoices").append(`<option value="${key}">${key}</option>`);
			}
		} else {
			setMessages("danger", e.message);
		}
	})
	.fail(function (e) {
		console.log(e);
	});
}

function addVotersFromList() {

	let list = $("#listChoices").val();
	let voters = votersLists[list];

	// $("#votersList").html("");
	for (let [email, procuration] of Object.entries(voters)) {
		addVoter();
		$("#votersList input[type='email']").last().val(email);
		$("#votersList input[type='number']").last().val(procuration - 1);
	}

}

function addChoicesFromList() {

	let list = $("#listChoicesChoix").val();
	let choices = choicesLists[list];

	// $("#votersList").html("");
	for (let i = 0; i < choices.length; i++) {
		addChoice();
		$("#choicesList input").last().val(choices[i]);
	}

}

// 		Ajout unique
let numChoice = 1;
function addChoice() {
	let choiceName = "choice" + numChoice;
	numChoice = numChoice + 1;

	$("#choicesList").append(`
	<div class="field has-addons" id="${choiceName}">
		<div class="control">
			<input class="input" type="text" name="${choiceName}" placeholder="Choix...">
		</div>
		<div class="control">
			<button class="button is-danger is-light" type="button" onclick="removeID('${choiceName}')">Retirer le choix</button>
		</div>
	</div>
`);
}

let numVoter = 1;
function addVoter() {
	let voterName = "voter" + numVoter;
	numVoter = numVoter + 1;

	$("#votersList").append(`
		<div class="field has-addons" id="${voterName}">
			<div class="control">
				<input class="input" type="email" placeholder="Email...">
			</div>
			<div class="control">
				<input class="input" type="number" placeholder="Procuration... [0-2]" value=0>
			</div>
			<div class="control">
				<button class="button is-danger is-light" type="button" onclick="removeID('${voterName}')">Retirer le votant</button>
			</div>
		</div>
	`);
}


// Consulter les scrutins
function consulter() {
	// Récupérer les filtres
	let filterVotable = $("#filterVotable").prop("checked");
	let filterResultats = $("#filterResultats").prop("checked");
	let filterCreateur = $("#filterCreateur").prop("checked");

	$.ajax({
			method: "GET",
			url: "ajax/scrutin/consulter.php",
			data: {},
			dataType: "json"
		})
		.done(function(e) {
			if (e.status == "success") {
			$("#scrutinsList").html("");
			
			// Inverser l'ordre
			e.votables.reverse();

			for (let i = 0; i < e.votables.length; i++) {
				// Vérifier les filtres
				if (filterVotable && (e.votables[i].votable == 0 || e.votables[i].remaining == 0 || e.votables[i].open == 0))
					continue;

				if (filterResultats && e.votables[i].open == 1)
					continue;

				if (filterCreateur && e.votables[i].manageable == 0)
					continue;

				let scrutin = e.votables[i];
				$("#scrutinsList").append(`
				<tr>
					<td>${scrutin.organizer}</td>
					<td>${scrutin.question}</td>
					<td>
						${scrutin.open ? "<span class='tag is-success'>Ouvert</span>" : "<span class='tag is-danger'>Fermé</span>"}
						<span class='tag is-${(scrutin.remaining == 0) ? "danger" : "success"}'>Votes restants : ${scrutin.remaining}</span>
						<span class='tag is-info'>Participation : ${(scrutin.totalUtilises)} / ${scrutin.totalPermis} (${parseInt(scrutin.totalUtilises/scrutin.totalPermis*100)}%)</span>
					</td>
					<td>
						${scrutin.votable ? "<span class='tag is-link is-light'>Votant</span>" : "<span class='tag is-danger is-light'>Non-Votant</span>"}
						${scrutin.manageable ? "<span class='tag is-success is-light'>Créateur</span>" : "<span class='tag is-danger is-light'>Non-Créateur</span>"}
					</td>
					<td>
						<button class='button is-info mr-1' onclick='fetchScrutin(\"${e.votables[i].id}\")' id="vote-${scrutin.id}">Voter</button>
						<button class='button is-success mr-1' onclick='resultatsScrutin(\"${e.votables[i].id}\")' id="resultats-${scrutin.id}">Résultats</button>
						<button class='button is-warning mr-1' onclick='cloturer(\"${e.votables[i].id}\")' id="cloturer-${scrutin.id}">Cloturer</button>
						<button class='button is-danger' onclick='detruire(\"${e.votables[i].id}\")' id="detruire-${scrutin.id}">Détruire</button>

					</td>
				</tr>
				`);
				// Voter
				if (scrutin.open == 0 || scrutin.votable == 0 || scrutin.remaining == 0)
					$(`#vote-${scrutin.id}`).prop("disabled", true);

				// Cloturer
				if (scrutin.open == 0 || scrutin.manageable == 0)
					$(`#cloturer-${scrutin.id}`).prop("disabled", true);

				// Résultats
				if (scrutin.open == 1)
					$(`#resultats-${scrutin.id}`).prop("disabled", true);

				// Détruire
				if (scrutin.manageable == 0)
					$(`#detruire-${scrutin.id}`).prop("disabled", true);
			}
		
			// setMessages(null, null);
		}
		else {
			setMessages("danger", e.message);
		}
		})
		.fail(function(e) {
			console.log(e);
		});
}

// 		Cloturer un scrutin
function cloturer(scrutinId) {
	$.ajax({
			method: "GET",
			url: "ajax/scrutin/cloturer.php",
			data: {
				"scrutinId": scrutinId,
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				publierResultats(scrutinId, e.data.choices, e.data.votes, e.data.systemeVote);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// 		Publier les résultats
function publierResultats(scrutinId, choices, votes, systemeVote) {
	let privateKey = localStorage.getItem(`scrutin-${scrutinId}`);
	if (privateKey == null) {
            setMessages("danger", i18n("private_key_missing"));
		return;
	}

	let encrypt = new JSEncrypt();
	encrypt.setPrivateKey(privateKey);

	// On déchiffre les votes
	for (let i = 0; i < votes.length; i++) {
		votes[i] = encrypt.decrypt(votes[i]);
	}

	// On compte les votes
	let results = {};

	if (systemeVote == "uninominal") {
		for (let i = 0; i < choices.length; i++) {
			results[choices[i]] = 0;
		}

		for (let i = 0; i < votes.length; i++) {
			if (results[votes[i]] >= 0)
				results[votes[i]]++;
		}

		console.log(results);

		// Sort results by occurences
		results = Object.fromEntries(
			Object.entries(results).sort(([,a],[,b]) => b-a)
		);

		console.log(results);



	} else if (systemeVote == "jugementMajoritaire") {
		for (let i = 0; i < choices.length; i++) {
			results[choices[i]] = {
				"repartition": Array(6).fill(0),
				"mention" : "NSP"
			};
		}

		for (let i = 0; i < votes.length; i++) {	
			let vote = JSON.parse(votes[i]);

			for (let [voteChoix, voteJugement] of Object.entries(vote)) {
				if (voteJugement != -1)
					results[voteChoix]["repartition"][voteJugement]++;
			}
		}

		// Calculate mentions, we look at 50% and say it's the mention
		let mentions = ["Très Bien", "Bien", "Assez Bien", "Passable", "Inssufisant", "À Rejeter", "NSP"];

		for (let [mention, value] of Object.entries(results)) {
			repartition = value["repartition"];

			// point médian of the repartition
			let sumRepartition = repartition.reduce((a, b) => a + b, 0);
			let median = Math.floor(sumRepartition / 2);

			let sum = 0;
			for (let i = 0; i < 6; i++) {
				sum += repartition[i];
				if (sum > median) {
					results[mention]["mention"] = mentions[i];
					break;
				}
			}
		}

		// Sort results by mentions
		results = Object.fromEntries(
			Object.entries(results).sort(([,a],[,b]) => mentions.indexOf(a["mention"]) - mentions.indexOf(b["mention"]))
		);

	}

	$.ajax({
		method: "GET",
		url: "ajax/scrutin/publierResultats.php",
		data: {
			"scrutinId": scrutinId,
			"resultats": JSON.stringify(results),
		},
		dataType: "json"
	})
	.done(function (e) {
		if (e.status == "success") {
			seeScrutins("consulterScrutins");
                        setMessages("warning", i18n("ballot_closed"));
			
			localStorage.removeItem(`scrutin-${scrutinId}`);
			sendNotification("resultats", scrutinId);
		} else {
			setMessages("danger", e.message);
		}
	})
	.fail(function (e) {
		console.log(e);
	});
}

// 		Détruire un scrutin
function detruire(scrutinId) {
	// Confirmation
	if (!confirm("Êtes-vous sûr de vouloir détruire ce scrutin ?"))
		return;

	$.ajax({
			method: "GET",
			url: "ajax/scrutin/detruire.php",
			data: {
				"scrutinId": scrutinId,
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				setMessages("success", e.message);
				consulter();
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// Gestion Listes
function getLists() {
	$.ajax({
		method: "GET",
		url: "ajax/liste/getLists.php",
		data: {},
		dataType: "json"
	})
	.done(function (e) {
		if (e.status == "success") {
			let data = e.data;
			$("#listeListes").html("");

			// for each key, value in data
			for (let [titreListe, votants] of Object.entries(data.lists)) {
				addListGerer(titreListe);

				// Add voters
				for (let [email, procuration] of Object.entries(votants)) {
					addVoterGerer(numList-1);
					$(`#voterGerer${numList-1}-${numVoterGerer-1} input[type='email']`).val(email);
					$(`#voterGerer${numList-1}-${numVoterGerer-1} input[type='number']`).val(procuration - 1);
				}
			}
		} else {
			setMessages("danger", e.message);
		}
	})
	.fail(function (e) {
		console.log(e);
	});
}

// 		Ajout d'une liste gérée
let numList = 0;
function addListGerer(titreListe = "") {
	$("#listeListes").append(
		`
		<div class="listeVotants" id="liste-${numList}">
			<div class="field">
				<label class="label" for="listeGerer${numList}">Titre de la liste : </label>
				<div class="control">
					<input class="input" type="text" id="listeGerer${numList}" placeholder="Titre de la liste ... ?" value="${titreListe}" onchange="setLists()">
				</div>
				<p class="help">Titre de la liste.</p>
			</div>
			
			<div class="field">
				<label class="label" for="question">Votants : </label>
				<div id="votersListGerer${numList}"></div>
				<p class="help">Ces votants pourront voter au scrutin. Maximum 2 procurations.</p>
				<button class="button is-success is-light" type="button" onclick="addVoterGerer(${numList})">Ajouter un votant</button>
			</div>

			<div class="field">
				<button class="button is-danger is-light" type="button" onclick="removeID('liste-${numList}'); setLists()" id="listeButton">Retirer la liste</button>
			</div>
		
			<hr>
		</div>
		`);

		numList++;
}

// 		Ajout d'un votant à une liste gérée
let numVoterGerer = 1;
function addVoterGerer(numeroList) {
	let voterName = "voterGerer" + numeroList + "-" + numVoterGerer;
	numVoterGerer = numVoterGerer + 1;

	$("#votersListGerer"+numeroList).append(`
		<div class="field has-addons" id="${voterName}">
			<div class="control">
				<input class="input" type="email" placeholder="Email..." onchange="setLists()">
			</div>
			<div class="control">
				<input class="input" type="number" placeholder="Procuration... [0-2]" value=0 onchange="setLists()">
			</div>
			<div class="control">
				<button class="button is-danger is-light" type="button" onclick="removeID('${voterName}'); setLists()">Retirer le votant</button>
			</div>
		</div>
	`);
}

// 		Enregistrer les listes gérées
function setLists() {
	let lists = {};

	$(".listeVotants").each(function (index) {
		let titreListe = $(this).find("input[type='text']").val();
		let votants = {};

		$(this).find("input[type='email']").each(function (index) {
			let email = $(this).val();

			if (email != "") {
				let procuration = parseInt($(this).parent().parent().find("input[type='number']").val()) + 1;
				
				votants[email] = procuration;
			}
		});

		lists[titreListe] = votants;
	});

	$.ajax({
			method: "GET",
			url: "ajax/liste/setLists.php",
			data: {
				"lists": JSON.stringify(lists)
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				setMessages("success", e.message);
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}


// Voter
// 		Récuperer les informations du scrutin
function fetchScrutin(id) {
	let scrutinId = id;

	$.ajax({
			method: "GET",
			url: "ajax/scrutin/fetchScrutin.php",
			data: {
				"scrutinId": scrutinId
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				seeScrutins("voterScrutin");
				clearForms();

				let data = e.data;
				$("#voteMessage").html(data.who);
				$("#voteQuestion").html(data.question);
				$("#keyScrutin").val(data.publicKey);
				$("#systemeVoteScrutin").val(data.systemeVote);
				$("#idScrutin").val(id);

				let choices = data.choices;
				for (let i = choices.length - 1; i > 0; i--) {
					const j = Math.floor(Math.random() * (i + 1));
					[choices[i], choices[j]] = [choices[j], choices[i]];
				}

				if (data.systemeVote == "uninominal") {
					$("#choicesUninominal").css("display", "initial");
					$("#voteChoicesUninominal").append(`<option selected value="">...</option>`);
					for (let i = 0; i < choices.length; i++) {
						$("#voteChoicesUninominal").append(`<option value="${choices[i]}">${choices[i]}</option>`);
					}
				} else if (data.systemeVote == "jugementMajoritaire") {
					$("#choicesJugementMajoritaire").css("display", "initial");
					for (let i = 0; i < choices.length; i++) {
						$("#voteChoicesJugementMajoritaire").append(
						`
						<tr>
							<td>${choices[i]}</td>
							<td>
								<span class="tag is-white">
									<label for="-1-${i}">
										<input type="radio" name="choixJM-${i}" id="-1-${i}" value="-1" checked />
										NSP
									</label>
								</span>
								<span class="tag is-info">
									<label for="0-${i}">
										<input type="radio" name="choixJM-${i}" id="0-${i}" value="0" />
										Très Bien
									</label>
								</span>
								<span class="tag is-success">
									<label for="1-${i}">
										<input type="radio" name="choixJM-${i}" id="1-${i}" value="1" />
										Bien
									</label>
								</span>
								<span class="tag is-success is-light">
									<label for="2-${i}">
										<input type="radio" name="choixJM-${i}" id="2-${i}" value="2" />
										Assez Bien
									</label>
								</span>
								<span class="tag is-danger is-light">
									<label for="3-${i}">
										<input type="radio" name="choixJM-${i}" id="3-${i}" value="3" />
										Passable
									</label>
								</span>
								<span class="tag is-danger">
									<label for="4-${i}">
										<input type="radio" name="choixJM-${i}" id="4-${i}" value="4" />
										Insuffisant
									</label>
								</span>
								<span class="tag is-black">
									<label for="5-${i}">
										<input type="radio" name="choixJM-${i}" id="5-${i}" value="5" />
										À Rejeter
									</label>
								</span>
							</td>
						</tr>
					`);
					}
				}
				
			} else {
				setMessages("danger", e.message);
				seeScrutins("consulterScrutins");
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// 		Envoyer le vote
function voter() {
	let scrutinId = $("#idScrutin").val();

	let publicKey = $("#keyScrutin").val();
	let encrypt = new JSEncrypt();
	encrypt.setPublicKey(publicKey);

	let systemeVote = $("#systemeVoteScrutin").val();
	let choice = "";
	if (systemeVote == "uninominal") {
		choice = $("#voteChoicesUninominal").val();
		if (choice == "") {
                    setMessages("danger", i18n("please_choose_option"));
			return;
		}
		choice = encrypt.encrypt(choice);
	}
	else if (systemeVote == "jugementMajoritaire") {
		choice = {};
		$("#voteChoicesJugementMajoritaire tr").each(function (index) {
			let choiceName = $(this).find("td").first().text();
			let choiceValue = $(this).find("input[type='radio']:checked").val();
			choice[choiceName] = choiceValue;
		});
		

		choice = encrypt.encrypt(JSON.stringify(choice));
	}

	$.ajax({
			method: "GET",
			url: "ajax/scrutin/voter.php",
			data: {
				"scrutinId": scrutinId,
				"choice": choice
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				// window.location.href = "?message=Vote enregistré avec succès";
				seeScrutins("consulterScrutins");
                                setMessages("success", i18n("vote_recorded"));
				clearForms();
			} else {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// Récupérer les résultats d'un scrutin
function resultatsScrutin(id) {
	let scrutinId = id;
	clearForms();
	
	$.ajax({
			method: "GET",
			url: "ajax/scrutin/resultatsScrutin.php",
			data: {
				"scrutinId": scrutinId
			},
			dataType: "json"
		})
		.done(function (e) {
			if (e.status == "success") {
				let data = e.data;

				$("#resultats-id").html(id);
				$("#resultats-organizer").html(data.organizer);
				$("#resultats-question").html(data.question);
				$("#resultats-participation").html(`${data.totalUtilises} / ${data.totalPermis} (${parseInt(data.totalUtilises/data.totalPermis*100)}%)`);

				let results = data.resultats;
				
				if (data.systemeVote == "uninominal") {
					$("#tableResultatsUninominal").css("display", "initial");

					// Display results
					for (let [key, value] of Object.entries(results)) {

						$("#resultats-results-uninominal").append(`<tr><td>${key}</td><td>${value}</td><td>${data.totalUtilises == 0 ? 0 :  parseInt(value / data.totalUtilises *100)}%</td><td>${parseInt(value / data.totalPermis *100)}%</td></tr>`);
					}
				} else if (data.systemeVote == "jugementMajoritaire") {
					$("#tableResultatsJugementMajoritaire").css("display", "initial");

					// Display results
					for (let [key, value] of Object.entries(results)) {
						let sumRepartition = value["repartition"].reduce((a, b) => a + b, 0);
						let pourcentages = value["repartition"].map(x => parseInt(sumRepartition == 0 ? 0 : x / sumRepartition * 100));

						$("#resultats-results-jugementMajoritaire").append(
						`<tr>
							<td>${key}</td>
							<td>
								<div class="progress" style="display:flex">
									<div class="progress has-background-info" style="width:${pourcentages[0]}%" max="100"></div>
									<div class="progress has-background-success" style="width:${pourcentages[1]}%" max="100"></div>
									<div class="progress has-background-success-light" style="width:${pourcentages[2]}%" max="100"></div>
									<div class="progress has-background-danger-light" style="width:${pourcentages[3]}%" max="100"></div>
									<div class="progress has-background-danger" style="width:${pourcentages[4]}%" max="100"></div>
									<div class="progress has-background-black" style="width:${pourcentages[5]}%" max="100"></div>
								</div>
							</td>
							<td>${value["mention"]}</td>
						</tr>`);
					}
				}

				// On change de panneau
				seeScrutins("resultatsScrutin");
				
				// setMessages(null, null);
			} else {
				setMessages("danger", e.message);
				seeScrutins("consulterScrutins");
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}

// UTILS
// Messages
function setMessages(type, text) {
	$("#messages").html("");
	// in click one 
	if (text != null) {
        var notification = $(`
            <div class="notification is-${type} mx-6" style="display: none;">
				<button class="delete" onclick="$(this).parent().remove();"></button>
				${text}
            </div>
        `);
        
        // Append notification to messages container
        $("#messages").append(notification);

        // Fade in the notification
        notification.fadeIn();

        // Set timeout to fade out and remove the notification after 5 seconds
        setTimeout(function() {
            notification.fadeOut(function() {
                $(this).remove();
            });
        }, 5000);
    }
}

// Clear forms
function clearForms() {
	$("#email").val("");
	$("#password").val("");

	$("#registerButton").prop("disabled", true);
	$("#loginButton").prop("disabled", true);

	$("#question").val("");
	$("#choicesList").html("");
	$("#votersList").html("");
	$("#votersListGerer").html("");

	numChoice = 1;
	numVoter = 1;
	numVoterGerer = 1;

	$("#voteChoicesUninominal").html("");
	$("#voteChoicesJugementMajoritaire").html("");
	$("#voteMessage").html("");
	$("#voteQuestion").html("");
	$("#choicesUninominal").css("display", "none");
	$("#choicesJugementMajoritaire").css("display", "none");

	$("#resultats-id").html("");
	$("#resultats-organizer").html("");
	$("#resultats-question").html("");
	$("#resultats-results-uninominal").html("");
	$("#resultats-results-jugementMajoritaire").html("");
	$("#tableResultatsUninominal").css("display", "none");
	$("#tableResultatsJugementMajoritaire").css("display", "none");
}

// Retirer un id
function removeID(id) {
	$("#" + id).remove();
}


// notifications-pdf
function sendNotification(typeNotif, scrutinId) {
	// Envoyer les notifications
	$.ajax({
		method: "GET",
		url: "ajax/notifications-pdf/envoyerNotification.php",
		data: {
			"scrutinId": scrutinId,
			"typeNotif": typeNotif,
		},
		dataType: "json"
	})
	.done(function (e) {
		if (e.status == "success") {
			seeScrutins("consulterScrutins");
                        setMessages("success", i18n("notifications_sent", {type: typeNotif}));
		} else {
			setMessages("danger", e.message);
		}
	})
	.fail(function (e) {
		console.log(e);
	});
}

function pdfResultats() {
	let scrutinId = $("#resultats-id").html();

	$.ajax({
			method: "GET",
			url: "ajax/notifications-pdf/getPdfResultats.php",
			data: {
				"scrutinId": scrutinId
			},
			dataType: "json"
		})
		.done(function (e) {
			console.log(e);
			if (e.status == "success") {
				// Convert the decoded data into an array buffer
				var decodedPdfData = atob(e.data.pdf);
				var arrayBuffer = new ArrayBuffer(decodedPdfData.length);
				var uint8Array = new Uint8Array(arrayBuffer);
				for (var i = 0; i < decodedPdfData.length; i++) {
					uint8Array[i] = decodedPdfData.charCodeAt(i);
				}
				let blob = new Blob([arrayBuffer], { type: 'application/pdf' });
				let url = URL.createObjectURL(blob);
				window.open(url, '_blank');

				setMessages("success", e.message);
			}
			else if (e.status == "error") {
				setMessages("danger", e.message);
			}
		})
		.fail(function (e) {
			console.log(e);
		});
}