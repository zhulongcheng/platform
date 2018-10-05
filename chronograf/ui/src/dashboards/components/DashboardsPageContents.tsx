// Libraries
import React, {Component, MouseEvent} from 'react'
import _ from 'lodash'

// Components
import DashboardsTable from 'src/dashboards/components/DashboardsTable'
import VisualizationIconLine from 'src/dashboards/graphics/VisualizationIconLine'
import VisualizationIconGauge from 'src/dashboards/graphics/VisualizationIconGauge'
import VisualizationIconBar from 'src/dashboards/graphics/VisualizationIconBar'
import VisualizationIconStepPlot from 'src/dashboards/graphics/VisualizationIconStepPlot'
import VisualizationIconStacked from 'src/dashboards/graphics/VisualizationIconStacked'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {Dashboard} from 'src/types/v2'
import {Notification} from 'src/types/notifications'
import {Blues, Purples, Greens, Greys} from 'src/clockface'

interface Props {
  dashboards: Dashboard[]
  defaultDashboardLink: string
  onSetDefaultDashboard: (dashboardLink: string) => void
  onDeleteDashboard: (dashboard: Dashboard) => () => void
  onCreateDashboard: () => void
  onCloneDashboard: (
    dashboard: Dashboard
  ) => (event: MouseEvent<HTMLButtonElement>) => void
  onExportDashboard: (dashboard: Dashboard) => () => void
  notify: (message: Notification) => void
  searchTerm: string
}

@ErrorHandling
class DashboardsPageContents extends Component<Props> {
  public render() {
    const {
      onDeleteDashboard,
      onCloneDashboard,
      onExportDashboard,
      onCreateDashboard,
      defaultDashboardLink,
      onSetDefaultDashboard,
      searchTerm,
    } = this.props

    return (
      <div className="col-md-12">
        <DashboardsTable
          searchTerm={searchTerm}
          dashboards={this.filteredDashboards}
          onDeleteDashboard={onDeleteDashboard}
          onCreateDashboard={onCreateDashboard}
          onCloneDashboard={onCloneDashboard}
          onExportDashboard={onExportDashboard}
          defaultDashboardLink={defaultDashboardLink}
          onSetDefaultDashboard={onSetDefaultDashboard}
        />
        <div style={{display: 'flex', flexWrap: 'wrap'}}>
          <div
            style={{
              backgroundColor: Greys.Castle,
              width: '200px',
              height: '200px',
              padding: '10px',
              margin: '2px',
            }}
          >
            <VisualizationIconGauge
              colorA={Blues.Pool}
              colorB={Purples.Comet}
              colorC={Greens.Rainforest}
              colorGrey={Greys.Sidewalk}
              stroke={1.5}
            />
          </div>
          <div
            style={{
              backgroundColor: Greys.Castle,
              width: '200px',
              height: '200px',
              padding: '10px',
              margin: '2px',
            }}
          >
            <VisualizationIconLine
              colorA={Blues.Pool}
              colorB={Purples.Comet}
              colorC={Greens.Rainforest}
              stroke={1.5}
            />
          </div>
          <div
            style={{
              backgroundColor: Greys.Castle,
              width: '200px',
              height: '200px',
              padding: '10px',
              margin: '2px',
            }}
          >
            <VisualizationIconBar
              colorA={Blues.Pool}
              colorB={Purples.Comet}
              colorC={Greens.Rainforest}
              stroke={1.5}
            />
          </div>
          <div
            style={{
              backgroundColor: Greys.Castle,
              width: '200px',
              height: '200px',
              padding: '10px',
              margin: '2px',
            }}
          >
            <VisualizationIconStepPlot
              colorA={Blues.Pool}
              colorB={Purples.Comet}
              colorC={Greens.Rainforest}
              stroke={1.5}
            />
          </div>
          <div
            style={{
              backgroundColor: Greys.Castle,
              width: '200px',
              height: '200px',
              padding: '10px',
              margin: '2px',
            }}
          >
            <VisualizationIconStacked
              colorA={Blues.Pool}
              colorB={Purples.Comet}
              colorC={Greens.Rainforest}
              stroke={1.5}
            />
          </div>
        </div>
      </div>
    )
  }

  private get filteredDashboards(): Dashboard[] {
    const {dashboards, searchTerm} = this.props

    const matchingDashboards = dashboards.filter(d =>
      d.name.toLowerCase().includes(searchTerm.toLowerCase())
    )

    return _.sortBy(matchingDashboards, d => d.name.toLowerCase())
  }
}

export default DashboardsPageContents
