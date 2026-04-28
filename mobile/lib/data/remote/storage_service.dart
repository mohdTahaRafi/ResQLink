import 'package:firebase_storage/firebase_storage.dart';
import 'package:flutter/foundation.dart';
import 'package:image_picker/image_picker.dart';
import 'package:uuid/uuid.dart';

/// Handles uploading images/media to Firebase Storage.
class StorageService {
  final FirebaseStorage _storage = FirebaseStorage.instance;
  static const _uuid = Uuid();

  /// Upload an XFile (from ImagePicker) to Firebase Storage.
  /// Returns the public download URL on success, null on failure.
  /// Times out after 30 seconds to prevent infinite hangs.
  Future<String?> uploadReportImage(XFile file) async {
    try {
      final ext = file.name.split('.').last;
      final fileName = '${_uuid.v4()}.$ext';
      final ref = _storage.ref().child('reports/$fileName');

      // Read bytes (works on both web and mobile)
      final bytes = await file.readAsBytes();
      final metadata = SettableMetadata(
        contentType: 'image/$ext',
      );

      // Upload with timeout
      final uploadTask = ref.putData(bytes, metadata);
      final snapshot = await uploadTask.timeout(
        const Duration(seconds: 30),
        onTimeout: () {
          uploadTask.cancel();
          throw Exception('Upload timed out after 30 seconds');
        },
      );

      final downloadUrl = await snapshot.ref.getDownloadURL();
      return downloadUrl;
    } catch (e) {
      debugPrint('Storage upload error: $e');
      return null;
    }
  }
}
