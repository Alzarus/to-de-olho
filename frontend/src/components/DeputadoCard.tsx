import { useState } from 'react';
import { User, Building2, MapPin, Euro } from 'lucide-react';
import { Deputado } from './DeputadosPage';

interface DeputadoCardProps {
  deputado: Deputado;
  onClick: () => void;
  onVerDespesas?: () => void;
}

export default function DeputadoCard({ deputado, onClick, onVerDespesas }: DeputadoCardProps) {
  const [imageError, setImageError] = useState(false);

  return (
    <div
      className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md 
                 transition-shadow cursor-pointer focus-within:ring-4 focus-within:ring-blue-300"
      onClick={onClick}
      role="button"
      tabIndex={0}
      aria-label={`Ver detalhes de ${deputado.nome}, ${deputado.siglaPartido} - ${deputado.siglaUf}`}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onClick();
        }
      }}
    >
      <div className="flex items-center mb-4">
        {deputado.urlFoto && !imageError ? (
          <img
            src={deputado.urlFoto}
            alt={`Foto oficial de ${deputado.nome}`}
            className="w-16 h-16 rounded-full object-cover mr-4"
            onError={() => setImageError(true)}
          />
        ) : (
          <div className="w-16 h-16 bg-gray-200 rounded-full flex items-center justify-center mr-4">
            <User className="h-8 w-8 text-gray-400" aria-hidden="true" />
          </div>
        )}
        <div>
          <h3 className="font-semibold text-gray-900 text-base leading-tight">
            {deputado.nome}
          </h3>
          <div className="flex items-center text-base text-gray-800 mt-1">
            <Building2 className="h-3 w-3 mr-1" aria-hidden="true" />
            <span className="font-medium">{deputado.siglaPartido}</span>
            <MapPin className="h-3 w-3 ml-2 mr-1" aria-hidden="true" />
            <span className="font-medium">{deputado.siglaUf}</span>
          </div>
        </div>
      </div>
      
      <div className="text-base text-gray-800">
        <p><strong>Situação:</strong> {deputado.condicaoEleitoral}</p>
        {deputado.email && (
          <p className="mt-1 truncate"><strong>Email:</strong> {deputado.email}</p>
        )}
      </div>

      <button 
        className="mt-4 w-full bg-blue-700 text-white text-base font-medium py-3 px-4 rounded-md 
                   hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300 
                   transition-colors duration-200 flex items-center justify-center"
        onClick={(e) => {
          e.stopPropagation();
          onVerDespesas?.();
        }}
        aria-label={`Ver gastos parlamentares de ${deputado.nome}`}
      >
        <Euro className="h-4 w-4 mr-2" aria-hidden="true" />
        Ver Gastos
      </button>
    </div>
  );
}
