import json
from jsonschema import validate, ValidationError
import string
import sys, getopt

testSchema = {
    'type':'object',
    'properties': {
        'testInput':
        {
            'type':'object',
            'properties': {
                'request': {
                    'type':'object',
                    'properties': {
                        'oslo.message': {'type':'string'},
                    },
                    'required': ['oslo.message']
                },
            },
            'required':['request']
        },
    },
    'required': ['testInput']
}

osloSchema = {
    'type':'object',
    'properties': {
        'publisher_id': {'type':'string'},
        'payload':      {'type':'array'},
    },
    'required': ['publisher_id','payload']
}

payloadSchema = {
    'type':'object',
    'properties': {
        'counter_name':   {'type': 'string'},
        'resource_id':    {'type': 'string'},
        'counter_volume': {'type': 'number'},
    },
    'required': ['counter_name','resource_id','counter_volume']
}


def generateResultsFromJSON(data):
    try:
        validate(instance=data, schema=testSchema)
        osloMessageJSON = data['testInput']['request']['oslo.message']

        message = json.loads(osloMessageJSON)
        validate(instance=message,schema=osloSchema)

        results = []
        for pl in message['payload']:
            validate(instance=pl,schema=payloadSchema)

            pluginAttr = pl["counter_name"].split('.')
            pt = ''
            typeInstance = ''
            if len(pluginAttr) > 1:
                pt = pluginAttr[1] 
                if len(pluginAttr) > 2:
                    typeInstance = pluginAttr[2]
            else:
                pt = pluginAttr[0]
            
            pluginInstance = pl['resource_id']
            plugin = pluginAttr[0]
            publisher = message['publisher_id']

            metricGenerated = {
                'publisher': publisher,
                'plugin': plugin,
                'plugin_instance': pl['resource_id'],
                'type': pt,
                'values': [pl['counter_volume']],
                'name': plugin,
                'key': publisher,
                'item_key': genItemKey(plugin, pluginInstance,pt, typeInstance),
                'type_instance': typeInstance,
                'labels': genLabels(pl, plugin, typeInstance, pluginInstance,publisher),
                'description': genDescription(pl, plugin, pt),
                'metric_name': genMetricName(plugin, pt, typeInstance),
            }

            results.append(metricGenerated)
        return results

    except ValidationError as e:
            print(e)
            print("---------")
            print(e.absolute_path)
        
            print("---------")
            print(e.absolute_schema_path)

def genMetricName(plugin, typ, typeInstance):
    name = ["ceilometer", plugin]
    if plugin != typ:
        name.append(typ)
    if typeInstance:
        name.append(typeInstance)
    return '_'.join(name)


def genDescription(payload, plugin, typ):
    id = payload["counter_name"]
    dstype = 'counter'
    if 'counter_type' in payload:
        dstype = payload['counter_type']
    return "Service Telemetry exporter: '{plugin}' Type: '{Type}' Dstype: '{Dstype}' Dsname: '{ID}'".format(plugin=plugin, Type=typ, Dstype=dstype, ID=id)


def genItemKey(plugin, pluginInstance, tpe, typeInstance):
    parts = [plugin]

    if plugin != tpe:
        parts.append(tpe)
    if pluginInstance:
        parts.append(pluginInstance)
    if typeInstance:
        parts.append(typeInstance)
    return "_".join(parts)

def genLabels(payload, plugin, typeInstance, pluginInstance, publisher):
    labels = {"publisher": publisher}
    if typeInstance:
        labels[plugin] = typeInstance
    else:
        labels[plugin] = pluginInstance
    
    if "counter_type" in payload:
        labels["type"] = payload["counter_type"]
    else:
        labels["type"] = "base"
    if "project_id" in payload:
        labels["project"] = payload["project_id"]
    if "resource_id" in payload:
        labels["resource"] = payload["resource_id"]
    if "counter_unit" in payload:
        labels["unit"] = payload["counter_unit"]
    if "counter_name" in payload:
        labels["counter"] = payload["counter_name"]
    return labels

def main(cmd, argv):
    inputFile = ''
    testName = ''

    try:
        opts, args = getopt.getopt(argv,"f:t:",["file=","test="])

        for opt, arg in opts:
            if opt in ('-f','--file'):
                inputFile = arg
            elif opt in ('-t','--test'):
                testName = arg

        with open(inputFile,'r') as jsonFile:
            data = json.load(jsonFile)
            if testName in data:
                data[testName]["validatedResults"] = generateResultsFromJSON(data[testName])
                print(json.dumps(data, indent=2))
            else:
                print("Could not find test '%s' in %s" % (testName,inputFile))

    except getopt.GetoptError:
        print("%s -f <filepath> -t <testname>" % cmd)


if __name__ == "__main__":
    main(sys.argv[0],sys.argv[1:])