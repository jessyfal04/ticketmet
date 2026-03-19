<!DOCTYPE html>
<html lang="fr">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link rel="stylesheet" href="lib/bulma.min.css">
	<link rel="stylesheet" href="lib/style.css">
	<script src="lib/jquery.min.js"></script>
	<script src="lib/script.js"></script>
        <script src="lib/jsencrypt.min.js"></script>
        <script type="module" src="lib/notificationModule.js"></script>
        <script src="lib/lang.js"></script>
	
	<link rel="shortcut icon" href="lib/icon.png" type="image/x-icon">
	<title>AVotey</title>
</head>
<body>
	<!-- Ouverture de session -->
	<?php
		session_start();
		$estConnecte = isset($_SESSION['uuid']);
	?>

	<!-- Hero de bienvenu -->
        <section class="hero is-info">
                <div class="hero-body">
                        <div class="field is-pulled-right">
                                <div class="control">
                                        <div class="select">
                                                <select id="langSelector" onchange="setLang(this.value)">
                                                        <option value="fr">Français</option>
                                                        <option value="en">English</option>
                                                        <option value="kr">한국어</option>
                                                </select>
                                        </div>
                                </div>
                        </div>
                        <p class="title" data-i18n="welcome">
                        Bienvenue sur AVotey
                        </p>
                        <p class="subtitle" data-i18n="subtitle">
                        Une plateforme de scrutins sécurisés
                        </p>
                </div>
        </section>

	<br>

	<div id="messages" style="position:sticky; top:0px; z-index:10;"></div>
	<br>

	<!-- Block de connexion -->
        <div class='box mx-6'>
                <h2 class="title is-2" data-i18n="account">Compte</h2>

		<!-- Texte de bienvenue / de demande de connection -->
                <p class='is-size-5 if-login' data-i18n="greeting" <?php if(!$estConnecte) {echo "style='display:none;'";} ?>>Bonjour <strong id="txt-connexion-uuid"><?php if($estConnecte) {echo $_SESSION['uuid'];} ?></strong></p>
                <p class='is-size-5 if-not-login' data-i18n="please_login" <?php if($estConnecte) {echo "style='display:none;'";} ?>>Veuillez vous connecter.</p>

		<br>

		<!-- Bouton de déconnexion -->
		<div class="block">
                        <button class="button is-danger is-medium if-login" onclick="logout()" <?php if(!$estConnecte) {echo "style='display:none;'";} ?> data-i18n="logout">Se déconnecter</button>
		</div>

		<!-- Boutons de notifications -->
		<div class="block">
                        <button class="button is-info is-medium if-login" id="activerNotifs" onclick="activerNotifications()" <?php if(!$estConnecte) {echo "style='display:none;'";} ?> data-i18n="enable_notifications">Activer Notifications</button>
                        <button class="button is-warning is-medium if-login" onclick="desactiverNotifications()" <?php if(!$estConnecte) {echo "style='display:none;'";} ?> data-i18n="disable_notifications">Désactiver Notifications</button>
		</div>
		

		<!-- Formulaire de connexion / inscription -->
		<div class="if-not-login" <?php if($estConnecte) {echo "style='display:none;'";} ?>>
			<div class="field">
                                <label class="label" for="email" data-i18n="email">Email : </label>
				<div class="control">
                                        <input class="input" type="email" name="email" id="email" placeholder="abc@xyz.fr" data-i18n-placeholder="email_placeholder" onchange="emailExists()" required autocomplete>
				</div>
                                <p class="help" data-i18n="valid_email">Email valide</p>
			</div>

			<div class="field">
                                <label class="label" for="password" data-i18n="password">Password : </label>
				<div class="control">
                                        <input class="input" type="password" name="password" id="password" minlength="8" placeholder="******" data-i18n-placeholder="password_placeholder" required>
				</div>
                                <p class="help" data-i18n="min_characters">Minimum 8 caractères</p>
			</div>

			<div class="field is-grouped">
				<div class="control mr-1">
                                        <button class="button is-link is-medium" type="button" onclick="login()" id="loginButton" disabled data-i18n="login">Login</button>
				</div>
				<div class="control">
                                        <button class="button is-link is-medium" type="button" onclick="register()" id="registerButton" disabled data-i18n="register">Register</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Blocs de scrutins -->
        <div class='box mx-6'>
                <h2 class="title is-2" data-i18n="scrutins">Scrutins</h2>

                <p class='is-size-5 if-not-login' id="txt-scrutins-notConnected" data-i18n="must_login_scrutins" <?php if($estConnecte) {echo "style='display:none;'";} ?>>Vous devez être connecté pour intéragir avec les scrutins</p>

		<div class="buttons block if-login" id="scrutinsButtons" <?php if(!$estConnecte) {echo "style='display:none;'";} ?>>
                        <button class="button is-link is-rounded is-medium is-outlined mr-1" onclick="seeScrutins('gererListes')" id="button-gererListes" data-i18n="gerer_listes">Gérer mes listes</button>
                        <button class="button is-link is-rounded is-medium is-outlined mr-1" onclick="seeScrutins('createScrutin')" id="button-createScrutin" data-i18n="create_scrutin">Créer un scrutin</button>
                        <button class="button is-link is-rounded is-medium is-outlined mr-1" onclick="seeScrutins('consulterScrutins')" id="button-consulterScrutins" data-i18n="voir_scrutins">Voir mes scrutins</button>

                        <button class="button is-info is-rounded is-medium" id="button-voterScrutin" style="display:none;" disabled data-i18n="voter_scrutin">Voter pour un structin</button>
                        <button class="button is-info is-rounded is-medium" id="button-resultatsScrutin" style="display:none;" disabled data-i18n="resultat_scrutin">Résultat pour un structin</button>
		</div>

		<div class="block" id="scrutins">

			<div id="gererListes" style="display:none;">
				<!-- <h5 class="title is-5">Gérer mes listes</h5> -->
				<div id="listeListes" ></div>
				<div class="field">
                                        <button class="button is-success is-light" type="button" onclick="addListGerer()" id="listeButton" data-i18n="add_list">Ajouter une liste</button>
				</div>
			</div>

			<div id="createScrutin" style="display:none;">
				<!-- <h5 class="title is-5">Créer un scrutin</h5> -->
				<div class="field">
                                        <label class="label" for="question" data-i18n="question">Question : </label>
					<div class="control">
                                                <input class="input" type="text" name="question" id="question" placeholder="Question ... ?" data-i18n-placeholder="question_placeholder">
					</div>
                                        <p class="help" data-i18n="question_help">Cette question sera posée aux votants.</p>
				</div>

				<div class="field">
                                        <label class="label" for="question" data-i18n="systeme_vote">Système de vote : </label>
					<div class="select is-rounded">
                                                <select id="systemeVote">
                                                        <option value="uninominal" default data-i18n="uninominal">Uninominal</option>
                                                        <option value="jugementMajoritaire" data-i18n="jugement_majoritaire">Jugement Majoritaire</option>
                                                </select>
					</div>
				</div>

				<div class="field">
                                        <label class="label" for="question" data-i18n="choice">Choix : </label>
					<div id="choicesList"></div>
                                        <p class="help" data-i18n="choices_help">Ces choix seront proposés aux votants.</p>
                                        <button class="button is-success is-light" type="button" onclick="addChoice()" data-i18n="add_choice">Ajouter un choix</button>
				</div>

				<div class="field">
                                        <label class="label" for="listChoicesChoix" data-i18n="predefined_choices">Choix prédéfinis : </label>
                                        <button class="button is-warning is-light" type="button" onclick="addChoicesFromList()" data-i18n="add_from_list">Ajouter depuis une liste</button>
					<div class="select is-rounded">
						<select id="listChoicesChoix">
                                                        <option value="" default data-i18n="option_default">...</option>
                                                        <option value="Binaire" data-i18n="binary">Binaire</option>
						</select>
					</div>
				</div>

				<div class="field">
                                        <label class="label" for="question" data-i18n="voters">Votants : </label>
					<div id="votersList"></div>
                                        <p class="help" data-i18n="voters_help">Ces votants pourront voter au scrutin. Maximum 2 procurations.</p>
                                        <button class="button is-success is-light" type="button" onclick="addVoter()" data-i18n="add_voter">Ajouter un votant</button>
				</div>

				<div class="field">
                                        <label class="label" for="question" data-i18n="liste_votants">Liste de votants : </label>
                                        <button class="button is-warning is-light" type="button" onclick="addVotersFromList()" data-i18n="add_from_list">Ajouter depuis une liste</button>
					<div class="select is-rounded">
						<select id="listChoices">
						</select>
					</div>
				</div>
				
				<div class="field">
                                        <button class="button is-success is-medium" type="button" onclick="create()" id="createButton" data-i18n="create_ballot">Créer le scrutin</button>
				</div>
			</div>

			<div id="consulterScrutins" style="display:none;">
				<!-- <h5 class="title is-5">Consulter mes scrutins</h5> -->
				<div class="block" onclick="consulter();">
                                        <h5 class="title is-5" data-i18n="filters">Filtres : </h5>
					<span class="tag is-medium is-info">
						<label class="checkbox">
							<input type="checkbox" id="filterVotable"/>
                                                        <span data-i18n="filter_votable">Votable</span>
						</label>
					</span>
					<span class="tag is-medium is-success">
						<label class="checkbox">
							<input type="checkbox" id="filterResultats"/>
                                                        <span data-i18n="filter_resultats">Résultats</span>
						</label>
					</span>
					<span class="tag is-medium is-link">
						<label class="checkbox">
							<input type="checkbox" id="filterCreateur"/>
                                                        <span data-i18n="filter_createur">Créateur</span>
						</label>
					</span>
				</div>

				<div class="block table-container">
                                        <h5 class="title is-5" data-i18n="list">Liste : </h3>
					<table class="table is-hoverable is-fullwidth">
						<thead>
							<tr>
                                                                <th data-i18n="organizer">Organisateur</th>
                                                                <th data-i18n="question_col">Question</th>
                                                                <th data-i18n="infos">Infos</th>
                                                                <th data-i18n="rights">Droits</th>
                                                                <th data-i18n="actions">Actions</th>
							</tr>
						</thead>
						<tbody id="scrutinsList">

						</tbody>
					</table>
				</div>
			</div>

			<div id="voterScrutin" style="display:none;">
				<!-- <h5 class="title is-5">Voter pour un scrutin</h5> -->
					
				<div class="field" style="">
                                        <label class="label" for="idScrutin" data-i18n="n_scrutin">N° Scrutin </label>
					<div class="control">
                                                <input class="input" type="text" name="idScrutin" id="idScrutin" placeholder="Veuillez passer par 'Voir mes scrutins' pour pouvoir voter" data-i18n-placeholder="placeholder_via_voir" disabled>
					</div>
					<p class="help">
                                                <span data-i18n="help_n_scrutin">Ceci représente le numéro du scrutin au quel vous allez voter.</span>
					</p>
				</div>

				<div class="field" style="display:none;">
                                        <label class="label" for="systemeVoteScrutin" data-i18n="systeme_vote_scrutin">Système de Vote</label>
					<div class="control">
                                                <input class="input" type="text" id="systemeVoteScrutin" placeholder="Veuillez passer par 'Voir mes scrutins' pour pouvoir voter" data-i18n-placeholder="placeholder_via_voir" disabled>
					</div>
                                        <p class="help" data-i18n="help_systeme_vote">Répresente le système de vote.</p>
				</div>

				<div class="field" style="">
                                        <label class="label" for="keyScrutin" data-i18n="public_key">Clé publique</label>
					<div class="control">
                                                <input class="input" type="text" name="keyScrutin" id="keyScrutin" placeholder="Veuillez passer par 'Voir mes scrutins' pour pouvoir voter" data-i18n-placeholder="placeholder_via_voir" disabled>
					</div>
                                        <p class="help" data-i18n="vote_encrypted">Le vote sera chiffré.</p>
				</div>

				<p class="is-4" id="voteMessage"></p>

				<hr>
				<h3 class="title is-3" id="voteQuestion">...</h3>

				<div class="field" id="choicesUninominal" style="display:none;">
                                        <label class="label" for="id" data-i18n="choice_uninominal">Choix Uninominal</label>
					<div class="select is-rounded">
						<select id="voteChoicesUninominal">
						</select>
					</div>
                                        <p class="help" data-i18n="choose_proposal">Veuillez choisir une proposition.</p>
				</div>

				<div class="field" id="choicesJugementMajoritaire" style="display:none;">
                                        <label class="label" for="id" data-i18n="choice_jugement">Choix Jugement Majoritaire</label>
					<div>
						<table class="table is-hoverable is-fullwidth">
							<thead>
								<tr>
									<th>Choix</th>
									<th>Jugement</th>
								</tr>
							</thead>
							<tbody id="voteChoicesJugementMajoritaire">
							</tbody>
						</table>
					</div>
                                        <p class="help" data-i18n="votez_proposition">Veuillez votez pour une proposition.</p>
				</div>
				
				<div class="field">
                                        <button class="button is-success is-medium" type="button" onclick="voter()" id="voterButton" data-i18n="vote_for_ballot">Voter pour le scrutin</button>
                                        <p class="help" data-i18n="vote_encrypted">Votre vote sera chiffré.</p>
				</div>
			</div>

			<div id="resultatsScrutin" class="printable" style="display:none;">
				<!-- <h4 class="title is-4">Résultat pour un scrutin</h5> -->

				<p class='is-size-5'>
					Pour le scrutin : <span id="resultats-id"></span><br>
					Organisateur : <span id="resultats-organizer"></span><br>
					La question était : <span id="resultats-question"></span><br>
					Participation : <span id="resultats-participation"></span><br>
				</p>

                                <h3 class="title is-3" data-i18n="results">Résultats :</h3>
				<table class="table is-hoverable is-fullwidth has-text-centered" id="tableResultatsUninominal" style="display:none;">
					<thead>
						<tr>
                                                        <th data-i18n="choice">Choix</th>
                                                        <th data-i18n="absolute">Absolu</th>
                                                        <th data-i18n="percent_voters">% Votants</th>
                                                        <th data-i18n="percent_inscrits">% Inscrits</th>
						</tr>
					</thead>
					<tbody id="resultats-results-uninominal">

					</tbody>
				</table>

				<table class="table is-hoverable is-fullwidth has-text-centered" id="tableResultatsJugementMajoritaire" style="display:none;">
					<thead>
						<tr>
                                                        <th data-i18n="choice">Choix</th>
                                                        <th data-i18n="repartition">Répartition</th>
                                                        <th data-i18n="mention">Mention</th>
						</tr>
					</thead>
					<tbody id="resultats-results-jugementMajoritaire">

					</tbody>
				</table>

				<div class="field">
                                        <button class="button is-info not-printable" type="button" onclick="pdfResultats()" data-i18n="download_pdf">Télécharger en PDF</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Footer -->
	<footer class="footer">
		<div class="content has-text-centered">
			<p>
                        <span data-i18n="footer_text"><strong>AVotey</strong> par Jessy FALLAVIER. Projet universitaire. Licence : <a href="http://creativecommons.org/licenses/by-nc-sa/4.0/">CC BY NC SA 4.0</a>.</span>
			</p>
		</div>
	</footer>
</body>
</html>