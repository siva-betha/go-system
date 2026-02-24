package machine_config

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	SaveMachines(ctx context.Context, machines []Machine) error
	GetMachines(ctx context.Context) ([]Machine, error)
}

type PgRepo struct {
	DB *pgxpool.Pool
}

var _ Repository = (*PgRepo)(nil)

func (r PgRepo) SaveMachines(ctx context.Context, machines []Machine) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clean up existing configuration for simplicity (full overwrite)
	// Alternatively, implement incremental updates
	_, err = tx.Exec(ctx, "DELETE FROM symbols")
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "DELETE FROM chambers")
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "DELETE FROM machines")
	if err != nil {
		return err
	}

	for _, m := range machines {
		_, err = tx.Exec(ctx,
			`INSERT INTO machines(id, name, ip, ams_net_id, port, created_at, updated_at)
			 VALUES($1, $2, $3, $4, $5, now(), now())`,
			m.ID, m.Name, m.IP, m.AmsNetID, m.Port,
		)
		if err != nil {
			return err
		}

		for _, c := range m.Chambers {
			_, err = tx.Exec(ctx,
				`INSERT INTO chambers(id, machine_id, name)
				 VALUES($1, $2, $3)`,
				c.ID, m.ID, c.Name,
			)
			if err != nil {
				return err
			}

			for _, s := range c.Symbols {
				_, err = tx.Exec(ctx,
					`INSERT INTO symbols(id, chamber_id, name, data_type, unit)
					 VALUES($1, $2, $3, $4, $5)`,
					s.ID, c.ID, s.Name, s.DataType, s.Unit,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func (r PgRepo) GetMachines(ctx context.Context) ([]Machine, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT m.id, m.name, m.ip, m.ams_net_id, m.port, m.created_at, m.updated_at,
		        c.id, c.name,
		        s.id, s.name, s.data_type, s.unit
		 FROM machines m
		 LEFT JOIN chambers c ON m.id = c.machine_id
		 LEFT JOIN symbols s ON c.id = s.chamber_id
		 ORDER BY m.id, c.id, s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machineMap := make(map[string]*Machine)
	chamberMap := make(map[string]*Chamber)

	for rows.Next() {
		var mID, mName, mIP, mNetID string
		var mPort int
		var mCreated, mUpdated time.Time
		var cID, cName *string
		var sID, sName, sType, sUnit *string

		err := rows.Scan(
			&mID, &mName, &mIP, &mNetID, &mPort, &mCreated, &mUpdated,
			&cID, &cName,
			&sID, &sName, &sType, &sUnit,
		)
		if err != nil {
			return nil, err
		}

		m, ok := machineMap[mID]
		if !ok {
			m = &Machine{
				ID:        mID,
				Name:      mName,
				IP:        mIP,
				AmsNetID:  mNetID,
				Port:      mPort,
				CreatedAt: mCreated,
				UpdatedAt: mUpdated,
				Chambers:  []Chamber{},
			}
			machineMap[mID] = m
		}

		if cID != nil {
			c, ok := chamberMap[*cID]
			if !ok {
				c = &Chamber{
					ID:        *cID,
					MachineID: mID,
					Name:      *cName,
					Symbols:   []Symbol{},
				}
				chamberMap[*cID] = c
				m.Chambers = append(m.Chambers, *c)
				// Re-fetch pointer because value copy was appended
				c = &m.Chambers[len(m.Chambers)-1]
				chamberMap[*cID] = c
			}

			if sID != nil {
				unit := ""
				if sUnit != nil {
					unit = *sUnit
				}
				s := Symbol{
					ID:        *sID,
					ChamberID: *cID,
					Name:      *sName,
					DataType:  *sType,
					Unit:      unit,
				}
				c.Symbols = append(c.Symbols, s)
			}
		}
	}

	out := make([]Machine, 0, len(machineMap))
	for _, m := range machineMap {
		out = append(out, *m)
	}
	return out, nil
}
