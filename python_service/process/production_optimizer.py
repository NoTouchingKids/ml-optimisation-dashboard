from process.base import BaseProcess
import gurobipy as gp
from gurobipy import GRB

# import pandas as pd
import numpy as np
import time
import json


class ProductionOptimizer(BaseProcess):
    def __init__(self, client_id: str, config: dict):
        super().__init__(client_id, config)
        self.model = None
        self.results = None

    def execute(self):
        try:
            self.log_status(
                "started", "Starting production optimization model", "optimize"
            )

            # Extract or generate data
            data = self.config.get("data", {})

            # Extract parameters from data or use defaults
            num_products = data.get("num_products", 10)
            num_resources = data.get("num_resources", 5)
            num_periods = data.get("num_periods", 6)

            self.logger.info(
                f"Model size: {num_products} products, {num_resources} resources, {num_periods} periods"
            )

            # Generate parameters if not provided
            if "demand" in data:
                demand = data["demand"]
            else:
                np.random.seed(42)
                demand = np.round(
                    np.random.uniform(50, 200, (num_products, num_periods))
                )
                self.logger.info("Generated random demand data")

            if "resource_capacity" in data:
                resource_capacity = data["resource_capacity"]
            else:
                resource_capacity = np.round(
                    np.random.uniform(800, 1500, (num_resources, num_periods))
                )
                self.logger.info("Generated random resource capacity data")

            if "resource_usage" in data:
                resource_usage = data["resource_usage"]
            else:
                resource_usage = np.round(
                    np.random.uniform(1, 10, (num_products, num_resources))
                )
                self.logger.info("Generated random resource usage data")

            if "production_cost" in data:
                production_cost = data["production_cost"]
            else:
                production_cost = np.round(np.random.uniform(10, 50, num_products))
                self.logger.info("Generated random production cost data")

            if "inventory_cost" in data:
                inventory_cost = data["inventory_cost"]
            else:
                inventory_cost = np.round(
                    production_cost * np.random.uniform(0.1, 0.3, num_products)
                )
                self.logger.info("Generated random inventory cost data")

            # Build the optimization model
            start_time = time.time()
            self.logger.info("Building Gurobi optimization model")

            model = gp.Model("ProductionOptimization")

            # Setup indices for the model
            products = range(num_products)
            resources = range(num_resources)
            periods = range(num_periods)

            # Create decision variables
            # Production quantity for each product in each period
            production = {}
            for p in products:
                for t in periods:
                    name = f"Produce_P{p}_T{t}"
                    production[p, t] = model.addVar(vtype=GRB.CONTINUOUS, name=name)

            # Inventory at the end of each period for each product
            inventory = {}
            for p in products:
                for t in periods:
                    name = f"Inventory_P{p}_T{t}"
                    inventory[p, t] = model.addVar(vtype=GRB.CONTINUOUS, name=name)

            model.update()
            self.logger.info(
                f"Created {model.NumVars} variables ({len(production)} production, {len(inventory)} inventory)"
            )

            # Add constraints
            constr_count = 0

            # 1. Inventory balance constraints
            for p in products:
                for t in periods:
                    if t == 0:
                        model.addConstr(
                            inventory[p, t] == production[p, t] - demand[p][t],
                            name=f"InventoryBalance_P{p}_T{t}",
                        )
                    else:
                        model.addConstr(
                            inventory[p, t]
                            == inventory[p, t - 1] + production[p, t] - demand[p][t],
                            name=f"InventoryBalance_P{p}_T{t}",
                        )
                    constr_count += 1

            # 2. Resource capacity constraints
            for r in resources:
                for t in periods:
                    model.addConstr(
                        gp.quicksum(
                            resource_usage[p][r] * production[p, t] for p in products
                        )
                        <= resource_capacity[r][t],
                        name=f"ResourceCapacity_R{r}_T{t}",
                    )
                    constr_count += 1

            # 3. Non-negative constraints
            for p in products:
                for t in periods:
                    model.addConstr(production[p, t] >= 0, name=f"NonNegProd_P{p}_T{t}")
                    model.addConstr(inventory[p, t] >= 0, name=f"NonNegInv_P{p}_T{t}")
                    constr_count += 2

            self.logger.info(f"Added {constr_count} constraints")

            # Set objective function: minimize total costs
            total_cost = gp.quicksum(
                production_cost[p] * production[p, t]
                + inventory_cost[p] * inventory[p, t]
                for p in products
                for t in periods
            )

            model.setObjective(total_cost, GRB.MINIMIZE)
            model.update()

            # Set solver parameters
            time_limit = self.config.get("time_limit", 30)
            mip_gap = self.config.get("mip_gap", 0.01)
            threads = self.config.get("threads", 0)

            model.setParam("TimeLimit", time_limit)
            model.setParam("MIPGap", mip_gap)
            model.setParam("Threads", threads)
            model.setParam("LogToConsole", 0)

            build_time = time.time() - start_time
            self.logger.info(f"Model built in {build_time:.2f} seconds")

            # Solve the model
            self.logger.info(f"Starting optimization (time_limit={time_limit}s)")
            solve_start = time.time()
            model.optimize()
            solve_time = time.time() - solve_start

            # Process results
            if model.status == GRB.OPTIMAL or model.status == GRB.TIME_LIMIT:
                self.logger.info(
                    f"Optimization complete. Status: {model.status}, Objective: {model.ObjVal:.2f}"
                )

                # Extract solution
                production_plan = {}
                inventory_plan = {}

                for p in products:
                    production_plan[f"Product_{p}"] = [
                        production[p, t].X for t in periods
                    ]
                    inventory_plan[f"Product_{p}"] = [
                        inventory[p, t].X for t in periods
                    ]

                # Calculate resource utilization
                resource_utilization = {}
                for r in resources:
                    resource_utilization[f"Resource_{r}"] = []
                    for t in periods:
                        usage = sum(
                            resource_usage[p][r] * production[p, t].X for p in products
                        )
                        capacity = resource_capacity[r][t]
                        utilization_pct = (
                            (usage / capacity) * 100 if capacity > 0 else 0
                        )
                        resource_utilization[f"Resource_{r}"].append(
                            {
                                "period": t,
                                "usage": float(usage),
                                "capacity": float(capacity),
                                "utilization_pct": float(utilization_pct),
                            }
                        )

                # Find bottlenecks (resources with highest utilization)
                bottlenecks = []
                for r in resources:
                    for t in periods:
                        usage = sum(
                            resource_usage[p][r] * production[p, t].X for p in products
                        )
                        capacity = resource_capacity[r][t]
                        util_pct = (usage / capacity) * 100 if capacity > 0 else 0
                        if (
                            util_pct > 95
                        ):  # Consider resources with >95% utilization as bottlenecks
                            bottlenecks.append(
                                {
                                    "resource": r,
                                    "period": t,
                                    "utilization": float(util_pct),
                                }
                            )

                # Calculate demand fulfillment
                demand_fulfillment = {}
                for p in products:
                    demand_fulfillment[f"Product_{p}"] = []
                    for t in periods:
                        demand_val = demand[p][t]
                        produced = production[p, t].X
                        demand_fulfillment[f"Product_{p}"].append(
                            {
                                "period": t,
                                "demand": float(demand_val),
                                "production": float(produced),
                            }
                        )

                # Calculate overall metrics
                total_demand = sum(sum(row) for row in demand)
                total_production = sum(
                    sum(production[p, t].X for t in periods) for p in products
                )
                total_inventory = sum(
                    sum(inventory[p, t].X for t in periods) for p in products
                )

                metrics = {
                    "objective_value": float(model.ObjVal),
                    "total_demand": float(total_demand),
                    "total_production": float(total_production),
                    "total_inventory": float(total_inventory),
                    "avg_inventory": float(
                        total_inventory / (num_products * num_periods)
                    ),
                    "bottleneck_count": len(bottlenecks),
                    "solve_time_seconds": solve_time,
                }

                self.logger.info(
                    json.dumps(
                        {
                            "optimization_metrics": metrics,
                            "bottlenecks": bottlenecks[:5] if bottlenecks else [],
                        }
                    )
                )

                # Prepare results
                results = {
                    "status": (
                        "optimal" if model.status == GRB.OPTIMAL else "time_limit"
                    ),
                    "objective_value": float(model.ObjVal),
                    "metrics": metrics,
                    "production_plan": production_plan,
                    "inventory_plan": inventory_plan,
                    "resource_utilization": resource_utilization,
                    "bottlenecks": bottlenecks,
                }

                self.log_status(
                    "completed",
                    f"Optimization completed with objective value: {model.ObjVal:.2f}",
                    "optimize",
                )
                return results

            else:
                status_text = {
                    GRB.INFEASIBLE: "Infeasible model",
                    GRB.UNBOUNDED: "Unbounded model",
                    GRB.INF_OR_UNBD: "Infeasible or unbounded",
                    GRB.NUMERIC: "Numeric issues",
                }.get(model.status, f"Failed with status {model.status}")

                self.logger.warning(f"Optimization failed: {status_text}")
                self.log_status(
                    "error", f"Optimization failed: {status_text}", "optimize"
                )

                return {
                    "status": "failed",
                    "status_text": status_text,
                    "error": f"Optimization failed with status {model.status}",
                }

        except gp.GurobiError as e:
            self.logger.error(f"Gurobi error: {str(e)}")
            self.log_status("error", f"Gurobi error: {str(e)}", "optimize")
            raise

        except Exception as e:
            self.logger.error(f"Unexpected error: {str(e)}")
            self.log_status("error", f"Optimization failed: {str(e)}", "optimize")
            raise
