# Golang Discord Bot

## Présentation

L'objectif de ce code est de déployer un bot discord qui automatise le processus de *Whitelist* sur un serveur **Minecraft** publique.

Lors de la réception d'un message dans le canal prédéfini et appelé par la balise ```!whitelist $PSEUDO```, notre BOT va requêter l'API de Minecraft pour vérifier que le pseudo renseigné existe bien, si celui-ci est correct alors le BOT se connectera au serveur de jeu via RCON et ajoutera le joueur à la whitelist.