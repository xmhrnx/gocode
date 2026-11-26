// Package remote lets the GoCode mobile app follow and steer the agent on
// a desktop. Only phones connect to desktops; desktops never reach out to
// phones or to each other.
//
// A phone is paired once with a Magic Wormhole code the desktop shows:
// the two swap public keys and the desktop's relay room. After that the
// phone finds the desktop on the local network, or, while the desktop has
// an agent working, through the relay Worker in relay/. Every connection
// is end-to-end encrypted and authenticated with the paired keys (Noise
// IK), so the relay only passes bytes along.
//
// The subpackages: wire is the messages, secure the encrypted channel,
// pair the pairing offer, host the desktop side and client the phone
// side.
package remote
