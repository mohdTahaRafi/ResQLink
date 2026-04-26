// File generated for Project SAMAJ (samaj-58742).
// Values extracted from google-services.json.
import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

class DefaultFirebaseOptions {
  static FirebaseOptions get currentPlatform {
    if (kIsWeb) {
      return web;
    }
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return android;
      case TargetPlatform.iOS:
        return ios;
      case TargetPlatform.macOS:
        return macos;
      case TargetPlatform.windows:
        return android; // fallback
      case TargetPlatform.linux:
        return android; // fallback
      case TargetPlatform.fuchsia:
        throw UnsupportedError('Fuchsia is not supported');
    }
  }

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'samaj-58742',
    storageBucket: 'samaj-58742.firebasestorage.app',
  );

  // TODO: Add iOS app in Firebase Console and fill these in
  static const FirebaseOptions ios = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'samaj-58742',
    storageBucket: 'samaj-58742.firebasestorage.app',
    iosBundleId: 'com.samaj.mobile',
  );

  // TODO: Add Web app in Firebase Console and fill these in
  static const FirebaseOptions web = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'samaj-58742',
    storageBucket: 'samaj-58742.firebasestorage.app',
    authDomain: 'samaj-58742.firebaseapp.com',
  );

  static const FirebaseOptions macos = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'samaj-58742',
    storageBucket: 'samaj-58742.firebasestorage.app',
    iosBundleId: 'com.samaj.mobile',
  );
}
