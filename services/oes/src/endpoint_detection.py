import numpy as np
from scipy import ndimage
from sklearn.decomposition import PCA
import logging

class OESEndpointDetector:
    def __init__(self, config):
        self.config = config
        self.logger = logging.getLogger(__name__)
        self.species_library = {
            'fluorine': [685.6, 703.7],
            'CF': [202.4, 208.2],
            'CF2': [251.9, 265.0],
            'CO': [483.5, 519.8],
            'silicon': [288.1, 390.5],
            'nitrogen': [337.1, 357.6],
            'argon': [750.4, 811.5],
            'oxygen': [777.2, 844.6],
            'chlorine': [837.6, 858.6],
        }

    def extract_wavelength(self, spectra, wavelength):
        intensities = []
        for spectrum in spectra:
            wl_array = np.array(spectrum['wavelengths'])
            int_array = np.array(spectrum['intensities'])
            intensity = np.interp(wavelength, wl_array, int_array)
            intensities.append(intensity)
        return np.array(intensities)

    def detect_endpoint_ratio(self, spectra, timestamps, signal_wavelength, reference_wavelength):
        signal_intensity = self.extract_wavelength(spectra, signal_wavelength)
        reference_intensity = self.extract_wavelength(spectra, reference_wavelength)
        
        ratio = signal_intensity / (reference_intensity + 1e-10)
        ratio_smooth = ndimage.gaussian_filter1d(ratio, sigma=2)
        derivative = np.gradient(ratio_smooth, timestamps)
        
        endpoint_idx = np.argmax(np.abs(derivative))
        return {
            'endpoint_time': float(timestamps[endpoint_idx]),
            'endpoint_idx': int(endpoint_idx),
            'confidence': float(np.abs(derivative[endpoint_idx]) / (np.mean(np.abs(derivative)) + 1e-10)),
            'ratio_values': ratio.tolist()
        }

    def detect_endpoint_multivariate(self, spectra, timestamps, wavelengths):
        n_wavelengths = len(wavelengths)
        n_times = len(timestamps)
        traces = np.zeros((n_times, n_wavelengths))
        
        for i, wl in enumerate(wavelengths):
            traces[:, i] = self.extract_wavelength(spectra, wl)
            
        pca = PCA(n_components=min(3, n_wavelengths))
        scores = pca.fit_transform(traces)
        
        # Simple T2-like change detection
        # Hotelling T2 would be better but let's stick to normalized derivative of first component
        pc1 = scores[:, 0]
        pc1_smooth = ndimage.gaussian_filter1d(pc1, sigma=3)
        derivative = np.gradient(pc1_smooth, timestamps)
        
        endpoint_idx = np.argmax(np.abs(derivative))
        return {
            'endpoint_time': float(timestamps[endpoint_idx]),
            'endpoint_idx': int(endpoint_idx),
            'explained_variance': pca.explained_variance_ratio_.tolist()
        }
