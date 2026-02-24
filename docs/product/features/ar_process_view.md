# Feature: Augmented Reality (AR) Process View

Mobile app that overlays real-time process data on the physical chamber using AR.

## User Story
As a maintenance technician, I want to point my phone at a chamber and see its current status, recent trends, and maintenance history.

## Implementation Details

### Identification & Tracking
- **Computer Vision**: Uses a lightweight CNN (Convolutional Neural Network) to identify tool models and unique serial numbers from the physical chassis.
- **Spatial Anchors**: Places persistent data "post-it notes" in 3D space around the chamber that stay fixed even when the user walks around.

### Hardware Integration
Connects to the API via WebSockets to provide <100ms latency for sensor data overlays.

```javascript
// mobile/ar/OverlayRenderer.js

const renderGauges = (toolData) => {
  return (
    <ARWorldSpace position={toolData.rfGeneratorCoord}>
      <Gauge 
        value={toolData.forwardPower} 
        unit="W" 
        label="RF Power"
        color={toolData.isOutofControl ? "red" : "green"}
      />
    </ARWorldSpace>
  );
};
```

## Killer Differentiator: Hands-Free Troubleshooting
- **"X-Ray vision"**: See through the stainless steel casing to visualize the internal plasma intensity or the current temperature of the ESC (Electrostatic Chuck).
- **Proactive Safety**: AR alerts flash red on surfaces that are hot to the touch or components with high-voltage risk.
- **Collaborative Remote Expert**: A senior engineer in another building can "draw" in the 3D space of the technician's phone to point out which valve to check.
